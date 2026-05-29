package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/VarvaraKurakova/fleet-events-api/internal/health"
	"github.com/VarvaraKurakova/fleet-events-api/internal/http/handlers"
)

func NewRouter(logger *slog.Logger, checker *health.Checker) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", handlers.Health)
	r.Get("/ready", handlers.Ready(checker))

	return r
}
