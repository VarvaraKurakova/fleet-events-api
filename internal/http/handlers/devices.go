package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/VarvaraKurakova/fleet-events-api/internal/domain"
	"github.com/VarvaraKurakova/fleet-events-api/internal/service"
)

type DeviceHandler struct {
	service *service.DeviceService
}

func NewDeviceHandler(service *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{
		service: service,
	}
}

type CreateDeviceRequest struct {
	VehicleID  string  `json:"vehicle_id"`
	ExternalID string  `json:"external_id"`
	Model      *string `json:"model"`
}

type DeviceResponse struct {
	ID         string  `json:"id"`
	VehicleID  string  `json:"vehicle_id"`
	ExternalID string  `json:"external_id"`
	Model      *string `json:"model,omitempty"`
	Status     string  `json:"status"`
	LastSeenAt *string `json:"last_seen_at,omitempty"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

type ListDevicesResponse struct {
	Items []DeviceResponse `json:"items"`
}

func (h *DeviceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request CreateDeviceRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	vehicleID, err := uuid.Parse(request.VehicleID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid vehicle id")
		return
	}

	device, err := h.service.Create(
		r.Context(),
		vehicleID,
		request.ExternalID,
		request.Model,
	)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toDeviceResponse(device))
}

func (h *DeviceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid device id")
		return
	}

	device, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toDeviceResponse(device))
}

func (h *DeviceHandler) List(w http.ResponseWriter, r *http.Request) {
	devices, err := h.service.List(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}

	items := make([]DeviceResponse, 0, len(devices))
	for _, device := range devices {
		items = append(items, toDeviceResponse(device))
	}

	writeJSON(w, http.StatusOK, ListDevicesResponse{
		Items: items,
	})
}

func toDeviceResponse(device domain.Device) DeviceResponse {
	var lastSeenAt *string
	if device.LastSeenAt != nil {
		formatted := device.LastSeenAt.Format("2006-01-02T15:04:05Z07:00")
		lastSeenAt = &formatted
	}

	return DeviceResponse{
		ID:         device.ID.String(),
		VehicleID:  device.VehicleID.String(),
		ExternalID: device.ExternalID,
		Model:      device.Model,
		Status:     device.Status,
		LastSeenAt: lastSeenAt,
		CreatedAt:  device.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  device.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
