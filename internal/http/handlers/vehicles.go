package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/VarvaraKurakova/fleet-events-api/internal/domain"
	"github.com/VarvaraKurakova/fleet-events-api/internal/service"
)

type VehicleHandler struct {
	service      *service.VehicleService
	stateService *service.VehicleStateService
}

func NewVehicleHandler(
	service *service.VehicleService,
	stateService *service.VehicleStateService,
) *VehicleHandler {
	return &VehicleHandler{
		service:      service,
		stateService: stateService,
	}
}

type CreateVehicleRequest struct {
	FleetID     string  `json:"fleet_id"`
	PlateNumber string  `json:"plate_number"`
	VIN         *string `json:"vin"`
	Type        string  `json:"type"`
}

type VehicleResponse struct {
	ID          string  `json:"id"`
	FleetID     string  `json:"fleet_id"`
	PlateNumber string  `json:"plate_number"`
	VIN         *string `json:"vin,omitempty"`
	Type        string  `json:"type"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type ListVehiclesResponse struct {
	Items []VehicleResponse `json:"items"`
}

type VehicleStateResponse struct {
	VehicleID    string   `json:"vehicle_id"`
	DeviceID     string   `json:"device_id"`
	EventID      string   `json:"event_id"`
	Lat          *float64 `json:"lat,omitempty"`
	Lon          *float64 `json:"lon,omitempty"`
	Speed        *float64 `json:"speed,omitempty"`
	BatteryLevel *float64 `json:"battery_level,omitempty"`
	Ignition     *bool    `json:"ignition,omitempty"`
	EventTime    string   `json:"event_time"`
	ReceivedAt   string   `json:"received_at"`
}

func (h *VehicleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request CreateVehicleRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	fleetID, err := uuid.Parse(request.FleetID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid fleet id")
		return
	}

	vehicle, err := h.service.Create(
		r.Context(),
		fleetID,
		request.PlateNumber,
		request.VIN,
		request.Type,
	)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toVehicleResponse(vehicle))
}

func (h *VehicleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid vehicle id")
		return
	}

	vehicle, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toVehicleResponse(vehicle))
}

func (h *VehicleHandler) List(w http.ResponseWriter, r *http.Request) {
	vehicles, err := h.service.List(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}

	items := make([]VehicleResponse, 0, len(vehicles))
	for _, vehicle := range vehicles {
		items = append(items, toVehicleResponse(vehicle))
	}

	writeJSON(w, http.StatusOK, ListVehiclesResponse{
		Items: items,
	})
}

func toVehicleResponse(vehicle domain.Vehicle) VehicleResponse {
	return VehicleResponse{
		ID:          vehicle.ID.String(),
		FleetID:     vehicle.FleetID.String(),
		PlateNumber: vehicle.PlateNumber,
		VIN:         vehicle.VIN,
		Type:        vehicle.Type,
		Status:      vehicle.Status,
		CreatedAt:   vehicle.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   vehicle.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *VehicleHandler) GetState(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid vehicle id")
		return
	}

	state, err := h.stateService.Get(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, VehicleStateResponse{
		VehicleID:    state.VehicleID.String(),
		DeviceID:     state.DeviceID.String(),
		EventID:      state.EventID.String(),
		Lat:          state.Lat,
		Lon:          state.Lon,
		Speed:        state.Speed,
		BatteryLevel: state.BatteryLevel,
		Ignition:     state.Ignition,
		EventTime:    state.EventTime.Format("2006-01-02T15:04:05Z07:00"),
		ReceivedAt:   state.ReceivedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
