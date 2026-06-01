package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/VarvaraKurakova/fleet-events-api/internal/apperrors"
	"github.com/VarvaraKurakova/fleet-events-api/internal/domain"
	"github.com/VarvaraKurakova/fleet-events-api/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type EventHandler struct {
	service *service.EventService
}

func NewEventHandler(service *service.EventService) *EventHandler {
	return &EventHandler{
		service: service,
	}
}

type CreateEventRequest struct {
	DeviceID     string                 `json:"device_id"`
	EventType    string                 `json:"event_type"`
	EventTime    string                 `json:"event_time"`
	Location     *EventLocationRequest  `json:"location"`
	Speed        *float64               `json:"speed"`
	BatteryLevel *float64               `json:"battery_level"`
	Ignition     *bool                  `json:"ignition"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type EventLocationRequest struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type CreateEventResponse struct {
	EventID string `json:"event_id"`
	Status  string `json:"status"`
}

type EventResponse struct {
	ID           string   `json:"id"`
	DeviceID     string   `json:"device_id"`
	VehicleID    string   `json:"vehicle_id"`
	EventType    string   `json:"event_type"`
	EventTime    string   `json:"event_time"`
	ReceivedAt   string   `json:"received_at"`
	Lat          *float64 `json:"lat,omitempty"`
	Lon          *float64 `json:"lon,omitempty"`
	Speed        *float64 `json:"speed,omitempty"`
	BatteryLevel *float64 `json:"battery_level,omitempty"`
	Ignition     *bool    `json:"ignition,omitempty"`
	CreatedAt    string   `json:"created_at"`
}

func (h *EventHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request CreateEventRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	eventTime, err := time.Parse(time.RFC3339, request.EventTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event_time")
		return
	}

	var lat *float64
	var lon *float64
	if request.Location != nil {
		lat = &request.Location.Lat
		lon = &request.Location.Lon
	}

	payload, err := json.Marshal(request)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	event, err := h.service.Ingest(
		r.Context(),
		service.IngestEventRequest{
			DeviceExternalID: request.DeviceID,
			EventType:        request.EventType,
			EventTime:        eventTime,
			Lat:              lat,
			Lon:              lon,
			Speed:            request.Speed,
			BatteryLevel:     request.BatteryLevel,
			Ignition:         request.Ignition,
			Payload:          payload,
		},
	)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusAccepted, toCreateEventResponse(event))
}

func (h *EventHandler) ListByVehicleID(w http.ResponseWriter, r *http.Request) {
	vehicleIDParam := chi.URLParam(r, "id")

	vehicleID, err := uuid.Parse(vehicleIDParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid vehicle id")
		return
	}

	events, err := h.service.ListByVehicleID(r.Context(), vehicleID)
	if err != nil {
		if errors.Is(err, apperrors.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "invalid input")
			return
		}

		writeError(w, http.StatusInternalServerError, "failed to list vehicle events")
		return
	}

	response := make([]EventResponse, 0, len(events))
	for _, event := range events {
		response = append(response, eventToResponse(event))
	}

	writeJSON(w, http.StatusOK, response)
}

func toCreateEventResponse(event domain.Event) CreateEventResponse {
	return CreateEventResponse{
		EventID: event.ID.String(),
		Status:  "accepted",
	}
}

func eventToResponse(event domain.Event) EventResponse {
	return EventResponse{
		ID:           event.ID.String(),
		DeviceID:     event.DeviceID.String(),
		VehicleID:    event.VehicleID.String(),
		EventType:    event.EventType,
		EventTime:    event.EventTime.Format(time.RFC3339),
		ReceivedAt:   event.ReceivedAt.Format(time.RFC3339),
		Lat:          event.Lat,
		Lon:          event.Lon,
		Speed:        event.Speed,
		BatteryLevel: event.BatteryLevel,
		Ignition:     event.Ignition,
		CreatedAt:    event.CreatedAt.Format(time.RFC3339),
	}
}
