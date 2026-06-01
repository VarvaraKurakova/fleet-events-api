package app

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/VarvaraKurakova/fleet-events-api/internal/config"
	"github.com/VarvaraKurakova/fleet-events-api/internal/messaging/rabbitmq"
	"github.com/VarvaraKurakova/fleet-events-api/internal/repository/postgres"
	"github.com/VarvaraKurakova/fleet-events-api/internal/service"
	workerpool "github.com/VarvaraKurakova/fleet-events-api/internal/worker"
)

func RunWorker(cfg config.Config, logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	postgresPool, err := pgxpool.New(ctx, cfg.Postgres.DSN)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer postgresPool.Close()

	if err := postgresPool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	rabbitMQ, err := amqp.Dial(cfg.RabbitMQ.URL)
	if err != nil {
		return fmt.Errorf("connect rabbitmq: %w", err)
	}
	defer rabbitMQ.Close()

	channel, err := rabbitMQ.Channel()
	if err != nil {
		return fmt.Errorf("open rabbitmq channel: %w", err)
	}
	defer channel.Close()

	if err := declareWorkerTopology(channel); err != nil {
		return err
	}

	if err := channel.Qos(
		cfg.Worker.Prefetch,
		0,
		false,
	); err != nil {
		return fmt.Errorf("set rabbitmq qos: %w", err)
	}

	deliveries, err := channel.Consume(
		rabbitmq.AlertsQueueName,
		"fleet-alerts-worker",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume alerts queue: %w", err)
	}

	alertRepository := postgres.NewAlertRepository(postgresPool)
	alertService := service.NewAlertService(alertRepository)

	pool := workerpool.NewPool(logger, alertService, cfg.Worker.Concurrency)
	pool.Start(ctx)
	defer pool.Stop()

	logger.Info(
		"worker started",
		"concurrency", cfg.Worker.Concurrency,
		"prefetch", cfg.Worker.Prefetch,
	)

	for {
		select {
		case <-ctx.Done():
			logger.Info("worker stopping")
			return nil

		case delivery, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("rabbitmq deliveries channel closed")
			}

			if err := pool.Submit(ctx, delivery); err != nil {
				logger.Error("failed to submit message to worker pool", "error", err)

				if nackErr := delivery.Nack(false, true); nackErr != nil {
					logger.Error("failed to nack message", "error", nackErr)
				}
			}
		}
	}
}

func declareWorkerTopology(channel *amqp.Channel) error {
	if err := channel.ExchangeDeclare(
		rabbitmq.EventsExchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare events exchange: %w", err)
	}

	if _, err := channel.QueueDeclare(
		rabbitmq.AlertsQueueName,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare alerts queue: %w", err)
	}

	if err := channel.QueueBind(
		rabbitmq.AlertsQueueName,
		rabbitmq.EventCreatedKey,
		rabbitmq.EventsExchangeName,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("bind alerts queue: %w", err)
	}

	return nil
}
