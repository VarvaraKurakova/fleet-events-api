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
	deviceAPIKey string,
	fleetHandler *handlers.FleetHandler,
	vehicleHandler *handlers.VehicleHandler,
	deviceHandler *handlers.DeviceHandler,
	eventHandler *handlers.EventHandler,
	alertHandler *handlers.AlertHandler,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logging(logger))
	r.Use(middleware.Recover(logger))

	r.Get("/health", handlers.Health)
	r.Get("/ready", handlers.Ready(checker))
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/fleets", fleetHandler.Create)
		r.Get("/fleets", fleetHandler.List)
		r.Get("/fleets/{id}", fleetHandler.GetByID)

		r.Post("/vehicles", vehicleHandler.Create)
		r.Get("/vehicles", vehicleHandler.List)
		r.Get("/vehicles/{id}/state", vehicleHandler.GetState)
		r.Get("/vehicles/{id}/events", eventHandler.ListByVehicleID)
		r.Get("/vehicles/{id}", vehicleHandler.GetByID)
		r.Post("/devices", deviceHandler.Create)
		r.Get("/devices", deviceHandler.List)
		r.Get("/devices/{id}", deviceHandler.GetByID)

		r.With(middleware.APIKey(deviceAPIKey)).Post("/events", eventHandler.Create)

		r.Get("/alerts", alertHandler.List)
		r.Patch("/alerts/{id}/resolve", alertHandler.Resolve)
	})

	return r
}
