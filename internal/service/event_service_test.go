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
	"github.com/VarvaraKurakova/fleet-events-api/internal/messaging"
)

type fakeEventRepositoryForEventService struct {
	createdEvent domain.Event
	createErr    error

	createCalled bool
}

func (r *fakeEventRepositoryForEventService) Create(
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
	r.createCalled = true

	if r.createErr != nil {
		return domain.Event{}, r.createErr
	}

	return r.createdEvent, nil
}

func (r *fakeEventRepositoryForEventService) GetLatestByVehicleID(
	ctx context.Context,
	vehicleID uuid.UUID,
) (domain.Event, error) {
	return domain.Event{}, nil
}

func (r *fakeEventRepositoryForEventService) ListByVehicleID(
	ctx context.Context,
	vehicleID uuid.UUID,
	filter domain.EventListFilter,
) ([]domain.Event, error) {
	return nil, nil
}

type fakeDeviceRepositoryForEventService struct {
	device domain.Device
	err    error

	getByExternalIDCalled bool
	updateLastSeenCalled  bool
	updatedLastSeenAt     time.Time
}

func (r *fakeDeviceRepositoryForEventService) GetByExternalID(
	ctx context.Context,
	externalID string,
) (domain.Device, error) {
	r.getByExternalIDCalled = true

	if r.err != nil {
		return domain.Device{}, r.err
	}

	return r.device, nil
}

func (r *fakeDeviceRepositoryForEventService) UpdateLastSeen(
	ctx context.Context,
	id uuid.UUID,
	lastSeenAt time.Time,
) error {
	r.updateLastSeenCalled = true
	r.updatedLastSeenAt = lastSeenAt

	return nil
}

func (r *fakeDeviceRepositoryForEventService) Create(
	ctx context.Context,
	vehicleID uuid.UUID,
	externalID string,
	model *string,
) (domain.Device, error) {
	return domain.Device{
		ID:         uuid.New(),
		VehicleID:  vehicleID,
		ExternalID: externalID,
		Model:      model,
		Status:     "active",
	}, nil
}

func (r *fakeDeviceRepositoryForEventService) List(
	ctx context.Context,
) ([]domain.Device, error) {
	return nil, nil
}

func (r *fakeDeviceRepositoryForEventService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Device, error) {
	return domain.Device{}, nil
}

type fakeVehicleStateCacheForEventService struct {
	setCalled bool
	setState  domain.VehicleState
	setErr    error
}

func (c *fakeVehicleStateCacheForEventService) Get(
	ctx context.Context,
	vehicleID uuid.UUID,
) (domain.VehicleState, error) {
	return domain.VehicleState{}, apperrors.ErrNotFound
}

func (c *fakeVehicleStateCacheForEventService) Set(
	ctx context.Context,
	state domain.VehicleState,
) error {
	c.setCalled = true
	c.setState = state

	return c.setErr
}

type fakeEventPublisherForEventService struct {
	publishCalled bool
	published     messaging.EventCreatedMessage
	err           error
}

func (p *fakeEventPublisherForEventService) PublishCreated(
	ctx context.Context,
	message messaging.EventCreatedMessage,
) error {
	p.publishCalled = true
	p.published = message

	return p.err
}

func TestEventService_Ingest_Success(t *testing.T) {
	deviceID := uuid.New()
	vehicleID := uuid.New()
	eventID := uuid.New()

	eventTime := time.Date(2026, 5, 11, 10, 15, 0, 0, time.UTC)
	receivedAt := time.Date(2026, 5, 11, 10, 15, 2, 0, time.UTC)

	lat := 55.7558
	lon := 37.6173
	speed := 73.4
	batteryLevel := 0.82
	ignition := true

	deviceRepository := &fakeDeviceRepositoryForEventService{
		device: domain.Device{
			ID:         deviceID,
			VehicleID:  vehicleID,
			ExternalID: "tracker-100500",
		},
	}

	eventRepository := &fakeEventRepositoryForEventService{
		createdEvent: domain.Event{
			ID:           eventID,
			DeviceID:     deviceID,
			VehicleID:    vehicleID,
			EventType:    "telemetry",
			EventTime:    eventTime,
			ReceivedAt:   receivedAt,
			Lat:          &lat,
			Lon:          &lon,
			Speed:        &speed,
			BatteryLevel: &batteryLevel,
			Ignition:     &ignition,
			Payload:      json.RawMessage(`{"source":"test"}`),
			CreatedAt:    receivedAt,
		},
	}

	stateCache := &fakeVehicleStateCacheForEventService{}
	publisher := &fakeEventPublisherForEventService{}

	service := NewEventService(
		eventRepository,
		deviceRepository,
		stateCache,
		publisher,
	)

	result, err := service.Ingest(context.Background(), IngestEventRequest{
		DeviceExternalID: "tracker-100500",
		EventType:        "telemetry",
		EventTime:        eventTime,
		Lat:              &lat,
		Lon:              &lon,
		Speed:            &speed,
		BatteryLevel:     &batteryLevel,
		Ignition:         &ignition,
		Payload:          json.RawMessage(`{"source":"test"}`),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.ID != eventID {
		t.Fatalf("expected event id %s, got %s", eventID, result.ID)
	}

	if !deviceRepository.getByExternalIDCalled {
		t.Fatalf("expected GetByExternalID to be called")
	}

	if !eventRepository.createCalled {
		t.Fatalf("expected EventRepository.Create to be called")
	}

	if !deviceRepository.updateLastSeenCalled {
		t.Fatalf("expected UpdateLastSeen to be called")
	}

	if !deviceRepository.updatedLastSeenAt.Equal(eventTime) {
		t.Fatalf("expected last_seen_at %s, got %s", eventTime, deviceRepository.updatedLastSeenAt)
	}

	if !stateCache.setCalled {
		t.Fatalf("expected VehicleStateCache.Set to be called")
	}

	if stateCache.setState.EventID != eventID {
		t.Fatalf("expected cached state event id %s, got %s", eventID, stateCache.setState.EventID)
	}

	if !publisher.publishCalled {
		t.Fatalf("expected publisher to be called")
	}

	if publisher.published.EventID != eventID.String() {
		t.Fatalf("expected published event id %s, got %s", eventID.String(), publisher.published.EventID)
	}
}

func TestEventService_Ingest_ReturnsInvalidInputForEmptyDeviceID(t *testing.T) {
	service := NewEventService(
		&fakeEventRepositoryForEventService{},
		&fakeDeviceRepositoryForEventService{},
		&fakeVehicleStateCacheForEventService{},
		&fakeEventPublisherForEventService{},
	)

	_, err := service.Ingest(context.Background(), IngestEventRequest{
		DeviceExternalID: "",
		EventType:        "telemetry",
		EventTime:        time.Now(),
		Payload:          json.RawMessage(`{}`),
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestEventService_Ingest_ReturnsErrorWhenPublisherFails(t *testing.T) {
	deviceID := uuid.New()
	vehicleID := uuid.New()
	eventID := uuid.New()

	eventTime := time.Date(2026, 5, 11, 10, 15, 0, 0, time.UTC)

	deviceRepository := &fakeDeviceRepositoryForEventService{
		device: domain.Device{
			ID:         deviceID,
			VehicleID:  vehicleID,
			ExternalID: "tracker-100500",
		},
	}

	eventRepository := &fakeEventRepositoryForEventService{
		createdEvent: domain.Event{
			ID:         eventID,
			DeviceID:   deviceID,
			VehicleID:  vehicleID,
			EventType:  "telemetry",
			EventTime:  eventTime,
			ReceivedAt: eventTime,
			Payload:    json.RawMessage(`{}`),
			CreatedAt:  eventTime,
		},
	}

	expectedErr := errors.New("publish failed")

	service := NewEventService(
		eventRepository,
		deviceRepository,
		&fakeVehicleStateCacheForEventService{},
		&fakeEventPublisherForEventService{
			err: expectedErr,
		},
	)

	_, err := service.Ingest(context.Background(), IngestEventRequest{
		DeviceExternalID: "tracker-100500",
		EventType:        "telemetry",
		EventTime:        eventTime,
		Payload:          json.RawMessage(`{}`),
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected publisher error, got %v", err)
	}
}
