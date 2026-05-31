package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/VarvaraKurakova/fleet-events-api/internal/apperrors"
	"github.com/VarvaraKurakova/fleet-events-api/internal/domain"
)

type EventRepository interface {
	Create(
		ctx context.Context,
		deviceID uuid.UUID,
		vehicleID uuid.UUID,
		eventType string,
		eventTime time.Time,
		lat *float64,
		lon *float64,
		speed *float64,
		batteryLevel *float64,
		ignition *bool,
		payload json.RawMessage,
	) (domain.Event, error)

	GetLatestByVehicleID(ctx context.Context, vehicleID uuid.UUID) (domain.Event, error)
}

type IngestEventRequest struct {
	DeviceExternalID string
	EventType        string
	EventTime        time.Time
	Lat              *float64
	Lon              *float64
	Speed            *float64
	BatteryLevel     *float64
	Ignition         *bool
	Payload          json.RawMessage
}

type VehicleStateCache interface {
	Set(ctx context.Context, state domain.VehicleState) error
	Get(ctx context.Context, vehicleID uuid.UUID) (domain.VehicleState, error)
}

type EventService struct {
	eventRepository  EventRepository
	deviceRepository DeviceRepository
	stateCache       VehicleStateCache
}

func NewEventService(
	eventRepository EventRepository,
	deviceRepository DeviceRepository,
	stateCache VehicleStateCache,
) *EventService {
	return &EventService{
		eventRepository:  eventRepository,
		deviceRepository: deviceRepository,
		stateCache:       stateCache,
	}
}

func (s *EventService) Ingest(ctx context.Context, request IngestEventRequest) (domain.Event, error) {
	request.DeviceExternalID = strings.TrimSpace(request.DeviceExternalID)
	request.EventType = strings.TrimSpace(request.EventType)

	if request.DeviceExternalID == "" || request.EventType == "" || request.EventTime.IsZero() {
		return domain.Event{}, apperrors.ErrInvalidInput
	}

	if request.Lat != nil && (*request.Lat < -90 || *request.Lat > 90) {
		return domain.Event{}, apperrors.ErrInvalidInput
	}

	if request.Lon != nil && (*request.Lon < -180 || *request.Lon > 180) {
		return domain.Event{}, apperrors.ErrInvalidInput
	}

	if request.BatteryLevel != nil && (*request.BatteryLevel < 0 || *request.BatteryLevel > 1) {
		return domain.Event{}, apperrors.ErrInvalidInput
	}

	if len(request.Payload) == 0 {
		request.Payload = json.RawMessage(`{}`)
	}

	device, err := s.deviceRepository.GetByExternalID(ctx, request.DeviceExternalID)
	if err != nil {
		return domain.Event{}, err
	}

	event, err := s.eventRepository.Create(
		ctx,
		device.ID,
		device.VehicleID,
		request.EventType,
		request.EventTime,
		request.Lat,
		request.Lon,
		request.Speed,
		request.BatteryLevel,
		request.Ignition,
		request.Payload,
	)
	if err != nil {
		return domain.Event{}, err
	}

	if err := s.deviceRepository.UpdateLastSeen(ctx, device.ID, event.EventTime); err != nil {
		return domain.Event{}, err
	}

	state := vehicleStateFromEvent(event)

	if err := s.stateCache.Set(ctx, state); err != nil {
		return domain.Event{}, err
	}

	return event, nil
}
