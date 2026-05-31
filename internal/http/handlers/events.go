package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/VarvaraKurakova/fleet-events-api/internal/domain"
	"github.com/VarvaraKurakova/fleet-events-api/internal/service"
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

func toCreateEventResponse(event domain.Event) CreateEventResponse {
	return CreateEventResponse{
		EventID: event.ID.String(),
		Status:  "accepted",
	}
}
