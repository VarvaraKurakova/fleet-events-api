package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/VarvaraKurakova/fleet-events-api/internal/health"
	"github.com/VarvaraKurakova/fleet-events-api/internal/http/handlers"
	"github.com/VarvaraKurakova/fleet-events-api/internal/http/middleware"
)

func NewRouter(
	logger *slog.Logger,
	checker *health.Checker,
	fleetHandler *handlers.FleetHandler,
	vehicleHandler *handlers.VehicleHandler,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logging(logger))

	r.Get("/health", handlers.Health)
	r.Get("/ready", handlers.Ready(checker))
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/fleets", fleetHandler.Create)
		r.Get("/fleets", fleetHandler.List)
		r.Get("/fleets/{id}", fleetHandler.GetByID)

		r.Post("/vehicles", vehicleHandler.Create)
		r.Get("/vehicles", vehicleHandler.List)
		r.Get("/vehicles/{id}", vehicleHandler.GetByID)
	})

	return r
}
