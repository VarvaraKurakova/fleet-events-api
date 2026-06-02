package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/VarvaraKurakova/fleet-events-api/internal/apperrors"
	"github.com/VarvaraKurakova/fleet-events-api/internal/domain"
	"github.com/VarvaraKurakova/fleet-events-api/internal/service"
)

type AlertHandler struct {
	service *service.AlertService
}

func NewAlertHandler(service *service.AlertService) *AlertHandler {
	return &AlertHandler{
		service: service,
	}
}

type AlertResponse struct {
	ID         string  `json:"id"`
	VehicleID  string  `json:"vehicle_id"`
	DeviceID   string  `json:"device_id"`
	EventID    *string `json:"event_id,omitempty"`
	Type       string  `json:"type"`
	Severity   string  `json:"severity"`
	Message    string  `json:"message"`
	Status     string  `json:"status"`
	CreatedAt  string  `json:"created_at"`
	ResolvedAt *string `json:"resolved_at,omitempty"`
}

func (h *AlertHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := getLimitOffset(r)

	var vehicleID *uuid.UUID
	vehicleIDParam := r.URL.Query().Get("vehicle_id")
	if vehicleIDParam != "" {
		parsedVehicleID, err := uuid.Parse(vehicleIDParam)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid vehicle_id")
			return
		}

		vehicleID = &parsedVehicleID
	}

	filter := domain.AlertListFilter{
		Status:    r.URL.Query().Get("status"),
		Type:      r.URL.Query().Get("type"),
		VehicleID: vehicleID,
		Limit:     limit,
		Offset:    offset,
	}

	alerts, err := h.service.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list alerts")
		return
	}

	response := make([]AlertResponse, 0, len(alerts))
	for _, alert := range alerts {
		response = append(response, alertToResponse(alert))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *AlertHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid alert id")
		return
	}

	alert, err := h.service.Resolve(r.Context(), id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			writeError(w, http.StatusNotFound, "alert not found")
			return
		}

		if errors.Is(err, apperrors.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "invalid input")
			return
		}

		writeError(w, http.StatusInternalServerError, "failed to resolve alert")
		return
	}

	writeJSON(w, http.StatusOK, alertToResponse(alert))
}

func alertToResponse(alert domain.Alert) AlertResponse {
	var eventID *string
	if alert.EventID != nil {
		value := alert.EventID.String()
		eventID = &value
	}

	var resolvedAt *string
	if alert.ResolvedAt != nil {
		value := alert.ResolvedAt.Format(time.RFC3339)
		resolvedAt = &value
	}

	return AlertResponse{
		ID:         alert.ID.String(),
		VehicleID:  alert.VehicleID.String(),
		DeviceID:   alert.DeviceID.String(),
		EventID:    eventID,
		Type:       alert.Type,
		Severity:   alert.Severity,
		Message:    alert.Message,
		Status:     alert.Status,
		CreatedAt:  alert.CreatedAt.Format(time.RFC3339),
		ResolvedAt: resolvedAt,
	}
}
