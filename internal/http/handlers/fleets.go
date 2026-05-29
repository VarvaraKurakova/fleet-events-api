package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/VarvaraKurakova/fleet-events-api/internal/apperrors"
	"github.com/VarvaraKurakova/fleet-events-api/internal/domain"
	"github.com/VarvaraKurakova/fleet-events-api/internal/service"
)

type FleetHandler struct {
	service *service.FleetService
}

func NewFleetHandler(service *service.FleetService) *FleetHandler {
	return &FleetHandler{
		service: service,
	}
}

type CreateFleetRequest struct {
	Name string `json:"name"`
}

type FleetResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ListFleetsResponse struct {
	Items []FleetResponse `json:"items"`
}

func (h *FleetHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request CreateFleetRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	fleet, err := h.service.Create(r.Context(), request.Name)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toFleetResponse(fleet))
}

func (h *FleetHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid fleet id")
		return
	}

	fleet, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toFleetResponse(fleet))
}

func (h *FleetHandler) List(w http.ResponseWriter, r *http.Request) {
	fleets, err := h.service.List(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}

	items := make([]FleetResponse, 0, len(fleets))
	for _, fleet := range fleets {
		items = append(items, toFleetResponse(fleet))
	}

	writeJSON(w, http.StatusOK, ListFleetsResponse{
		Items: items,
	})
}

func toFleetResponse(fleet domain.Fleet) FleetResponse {
	return FleetResponse{
		ID:        fleet.ID.String(),
		Name:      fleet.Name,
		CreatedAt: fleet.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: fleet.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, apperrors.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid input")
	case errors.Is(err, apperrors.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
