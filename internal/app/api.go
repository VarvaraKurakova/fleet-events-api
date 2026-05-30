package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"

	"github.com/VarvaraKurakova/fleet-events-api/internal/config"
	"github.com/VarvaraKurakova/fleet-events-api/internal/health"
	httptransport "github.com/VarvaraKurakova/fleet-events-api/internal/http"
	"github.com/VarvaraKurakova/fleet-events-api/internal/http/handlers"
	"github.com/VarvaraKurakova/fleet-events-api/internal/repository/postgres"
	"github.com/VarvaraKurakova/fleet-events-api/internal/service"
)

func RunAPI(cfg config.Config, logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	postgresPool, err := pgxpool.New(ctx, cfg.Postgres.DSN)
	if err != nil {
		return err
	}
	defer postgresPool.Close()

	if err := postgresPool.Ping(ctx); err != nil {
		return err
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       0,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return err
	}

	rabbitConn, err := amqp091.Dial(cfg.RabbitMQ.URL)
	if err != nil {
		return err
	}
	defer rabbitConn.Close()

	checker := health.NewChecker(postgresPool, redisClient, rabbitConn)

	fleetRepository := postgres.NewFleetRepository(postgresPool)
	fleetService := service.NewFleetService(fleetRepository)
	fleetHandler := handlers.NewFleetHandler(fleetService)

	vehicleRepository := postgres.NewVehicleRepository(postgresPool)
	vehicleService := service.NewVehicleService(vehicleRepository, fleetRepository)
	vehicleHandler := handlers.NewVehicleHandler(vehicleService)

	router := httptransport.NewRouter(logger, checker, fleetHandler, vehicleHandler)

	server := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)

	go func() {
		logger.Info("starting api server", "addr", cfg.HTTP.Addr)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}

		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}

		logger.Info("api server stopped gracefully")
		return nil

	case err := <-errCh:
		return err
	}
}
