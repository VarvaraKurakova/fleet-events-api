package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/VarvaraKurakova/fleet-events-api/internal/http/handlers"
)

func NewRouter(logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", handlers.Health)

	return r
}
