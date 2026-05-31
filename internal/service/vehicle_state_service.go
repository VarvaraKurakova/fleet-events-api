package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/VarvaraKurakova/fleet-events-api/internal/apperrors"
	"github.com/VarvaraKurakova/fleet-events-api/internal/domain"
)

type VehicleStateService struct {
	stateCache      VehicleStateCache
	eventRepository EventRepository
}

func NewVehicleStateService(stateCache VehicleStateCache, eventRepository EventRepository) *VehicleStateService {
	return &VehicleStateService{
		stateCache:      stateCache,
		eventRepository: eventRepository,
	}
}

func (s *VehicleStateService) Get(ctx context.Context, vehicleID uuid.UUID) (domain.VehicleState, error) {
	if vehicleID == uuid.Nil {
		return domain.VehicleState{}, apperrors.ErrInvalidInput
	}

	state, err := s.stateCache.Get(ctx, vehicleID)
	if err == nil {
		return state, nil
	}

	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return domain.VehicleState{}, err
	}

	event, err := s.eventRepository.GetLatestByVehicleID(ctx, vehicleID)
	if err != nil {
		return domain.VehicleState{}, err
	}

	state = vehicleStateFromEvent(event)

	if err := s.stateCache.Set(ctx, state); err != nil {
		return domain.VehicleState{}, err
	}

	return state, nil
}

func vehicleStateFromEvent(event domain.Event) domain.VehicleState {
	return domain.VehicleState{
		VehicleID:    event.VehicleID,
		DeviceID:     event.DeviceID,
		EventID:      event.ID,
		Lat:          event.Lat,
		Lon:          event.Lon,
		Speed:        event.Speed,
		BatteryLevel: event.BatteryLevel,
		Ignition:     event.Ignition,
		EventTime:    event.EventTime,
		ReceivedAt:   event.ReceivedAt,
	}
}
