package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/VarvaraKurakova/fleet-events-api/internal/health"
)

type ReadyResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

func Ready(checker *health.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := timeOutContext(r)
		defer cancel()

		checks := map[string]string{
			"postgres": "ok",
			"redis":    "ok",
			"rabbitmq": "ok",
		}

		statusCode := http.StatusOK
		status := "ready"

		if err := checker.CheckPostgres(ctx); err != nil {
			checks["postgres"] = "failed"
			statusCode = http.StatusServiceUnavailable
			status = "not_ready"
		}

		if err := checker.CheckRedis(ctx); err != nil {
			checks["redis"] = "failed"
			statusCode = http.StatusServiceUnavailable
			status = "not_ready"
		}

		if err := checker.CheckRabbitMQ(ctx); err != nil {
			checks["rabbitmq"] = "failed"
			statusCode = http.StatusServiceUnavailable
			status = "not_ready"
		}

		response := ReadyResponse{
			Status: status,
			Checks: checks,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		_ = json.NewEncoder(w).Encode(response)
	}
}

func timeOutContext(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), 2*time.Second)
}
