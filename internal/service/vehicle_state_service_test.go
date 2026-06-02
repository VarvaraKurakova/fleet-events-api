package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/VarvaraKurakova/fleet-events-api/internal/apperrors"
	"github.com/VarvaraKurakova/fleet-events-api/internal/domain"
)

type fakeVehicleStateCache struct {
	state domain.VehicleState
	err   error

	setCalled bool
	setState  domain.VehicleState
}

func (c *fakeVehicleStateCache) Get(
	ctx context.Context,
	vehicleID uuid.UUID,
) (domain.VehicleState, error) {
	return c.state, c.err
}

func (c *fakeVehicleStateCache) Set(
	ctx context.Context,
	state domain.VehicleState,
) error {
	c.setCalled = true
	c.setState = state

	return nil
}

type fakeEventRepositoryForVehicleState struct {
	event domain.Event
	err   error

	getLatestCalled bool
}

func (r *fakeEventRepositoryForVehicleState) Create(
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
) (domain.Event, error) {
	return domain.Event{}, nil
}

func (r *fakeEventRepositoryForVehicleState) GetLatestByVehicleID(
	ctx context.Context,
	vehicleID uuid.UUID,
) (domain.Event, error) {
	r.getLatestCalled = true

	return r.event, r.err
}

func (r *fakeEventRepositoryForVehicleState) ListByVehicleID(
	ctx context.Context,
	vehicleID uuid.UUID,
	filter domain.EventListFilter,
) ([]domain.Event, error) {
	return nil, nil
}

func TestVehicleStateService_Get_ReturnsStateFromCache(t *testing.T) {
	vehicleID := uuid.New()
	deviceID := uuid.New()
	eventID := uuid.New()

	state := domain.VehicleState{
		VehicleID: vehicleID,
		DeviceID:  deviceID,
		EventID:   eventID,
		EventTime: time.Now(),
	}

	cache := &fakeVehicleStateCache{
		state: state,
		err:   nil,
	}

	eventRepository := &fakeEventRepositoryForVehicleState{}

	service := NewVehicleStateService(cache, eventRepository)

	result, err := service.Get(context.Background(), vehicleID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.VehicleID != vehicleID {
		t.Fatalf("expected vehicle id %s, got %s", vehicleID, result.VehicleID)
	}

	if eventRepository.getLatestCalled {
		t.Fatalf("expected event repository not to be called on cache hit")
	}

	if cache.setCalled {
		t.Fatalf("expected cache Set not to be called on cache hit")
	}
}

func TestVehicleStateService_Get_FallbacksToLatestEventOnCacheMiss(t *testing.T) {
	vehicleID := uuid.New()
	deviceID := uuid.New()
	eventID := uuid.New()

	speed := 75.0
	batteryLevel := 0.80
	lat := 55.7558
	lon := 37.6173
	ignition := true

	event := domain.Event{
		ID:           eventID,
		VehicleID:    vehicleID,
		DeviceID:     deviceID,
		EventType:    "telemetry",
		EventTime:    time.Now(),
		ReceivedAt:   time.Now(),
		Lat:          &lat,
		Lon:          &lon,
		Speed:        &speed,
		BatteryLevel: &batteryLevel,
		Ignition:     &ignition,
	}

	cache := &fakeVehicleStateCache{
		err: apperrors.ErrNotFound,
	}

	eventRepository := &fakeEventRepositoryForVehicleState{
		event: event,
		err:   nil,
	}

	service := NewVehicleStateService(cache, eventRepository)

	result, err := service.Get(context.Background(), vehicleID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !eventRepository.getLatestCalled {
		t.Fatalf("expected event repository to be called on cache miss")
	}

	if !cache.setCalled {
		t.Fatalf("expected cache Set to be called after fallback")
	}

	if result.VehicleID != vehicleID {
		t.Fatalf("expected vehicle id %s, got %s", vehicleID, result.VehicleID)
	}

	if result.EventID != eventID {
		t.Fatalf("expected event id %s, got %s", eventID, result.EventID)
	}

	if result.Speed == nil || *result.Speed != speed {
		t.Fatalf("expected speed %.1f, got %v", speed, result.Speed)
	}

	if cache.setState.EventID != eventID {
		t.Fatalf("expected cached state event id %s, got %s", eventID, cache.setState.EventID)
	}
}

func TestVehicleStateService_Get_ReturnsInvalidInputForNilVehicleID(t *testing.T) {
	cache := &fakeVehicleStateCache{}
	eventRepository := &fakeEventRepositoryForVehicleState{}

	service := NewVehicleStateService(cache, eventRepository)

	_, err := service.Get(context.Background(), uuid.Nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
