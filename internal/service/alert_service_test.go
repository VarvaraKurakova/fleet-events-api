package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/VarvaraKurakova/fleet-events-api/internal/domain"
	"github.com/VarvaraKurakova/fleet-events-api/internal/messaging"
)

type fakeAlertRepository struct {
	createdAlerts []domain.Alert
}

func (r *fakeAlertRepository) Create(
	ctx context.Context,
	vehicleID uuid.UUID,
	deviceID uuid.UUID,
	eventID *uuid.UUID,
	alertType string,
	severity string,
	message string,
) (domain.Alert, error) {
	alert := domain.Alert{
		ID:        uuid.New(),
		VehicleID: vehicleID,
		DeviceID:  deviceID,
		EventID:   eventID,
		Type:      alertType,
		Severity:  severity,
		Message:   message,
		Status:    domain.AlertStatusOpen,
	}

	r.createdAlerts = append(r.createdAlerts, alert)

	return alert, nil
}

func (r *fakeAlertRepository) List(
	ctx context.Context,
	filter domain.AlertListFilter,
) ([]domain.Alert, error) {
	return r.createdAlerts, nil
}

func (r *fakeAlertRepository) Resolve(
	ctx context.Context,
	id uuid.UUID,
) (domain.Alert, error) {
	for _, alert := range r.createdAlerts {
		if alert.ID == id {
			alert.Status = domain.AlertStatusResolved
			return alert, nil
		}
	}

	return domain.Alert{}, nil
}

func TestAlertService_ProcessEvent_CreatesSpeedAlert(t *testing.T) {
	repository := &fakeAlertRepository{}
	service := NewAlertService(repository)

	speed := 110.5
	batteryLevel := 0.80
	lat := 55.7558
	lon := 37.6173

	message := newTestEventCreatedMessage(t, speed, batteryLevel, lat, lon)

	alerts, err := service.ProcessEvent(context.Background(), message)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}

	if alerts[0].Type != domain.AlertTypeSpeedLimitExceeded {
		t.Fatalf("expected alert type %q, got %q", domain.AlertTypeSpeedLimitExceeded, alerts[0].Type)
	}

	if alerts[0].Severity != domain.AlertSeverityWarning {
		t.Fatalf("expected severity %q, got %q", domain.AlertSeverityWarning, alerts[0].Severity)
	}
}

func TestAlertService_ProcessEvent_CreatesLowBatteryAlert(t *testing.T) {
	repository := &fakeAlertRepository{}
	service := NewAlertService(repository)

	speed := 60.0
	batteryLevel := 0.10
	lat := 55.7558
	lon := 37.6173

	message := newTestEventCreatedMessage(t, speed, batteryLevel, lat, lon)

	alerts, err := service.ProcessEvent(context.Background(), message)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}

	if alerts[0].Type != domain.AlertTypeLowBattery {
		t.Fatalf("expected alert type %q, got %q", domain.AlertTypeLowBattery, alerts[0].Type)
	}
}

func TestAlertService_ProcessEvent_CreatesTwoAlerts(t *testing.T) {
	repository := &fakeAlertRepository{}
	service := NewAlertService(repository)

	speed := 110.5
	batteryLevel := 0.10
	lat := 55.7558
	lon := 37.6173

	message := newTestEventCreatedMessage(t, speed, batteryLevel, lat, lon)

	alerts, err := service.ProcessEvent(context.Background(), message)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(alerts) != 2 {
		t.Fatalf("expected 2 alerts, got %d", len(alerts))
	}

	alertTypes := map[string]bool{}
	for _, alert := range alerts {
		alertTypes[alert.Type] = true
	}

	if !alertTypes[domain.AlertTypeSpeedLimitExceeded] {
		t.Fatalf("expected speed limit alert")
	}

	if !alertTypes[domain.AlertTypeLowBattery] {
		t.Fatalf("expected low battery alert")
	}
}

func TestAlertService_ProcessEvent_CreatesNoAlertsForNormalEvent(t *testing.T) {
	repository := &fakeAlertRepository{}
	service := NewAlertService(repository)

	speed := 60.0
	batteryLevel := 0.80
	lat := 55.7558
	lon := 37.6173

	message := newTestEventCreatedMessage(t, speed, batteryLevel, lat, lon)

	alerts, err := service.ProcessEvent(context.Background(), message)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts, got %d", len(alerts))
	}
}

func TestAlertService_ProcessEvent_CreatesInvalidLocationAlert(t *testing.T) {
	repository := &fakeAlertRepository{}
	service := NewAlertService(repository)

	speed := 60.0
	batteryLevel := 0.80

	message := newTestEventCreatedMessageWithoutLocation(t, speed, batteryLevel)

	alerts, err := service.ProcessEvent(context.Background(), message)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}

	if alerts[0].Type != domain.AlertTypeInvalidLocation {
		t.Fatalf("expected alert type %q, got %q", domain.AlertTypeInvalidLocation, alerts[0].Type)
	}
}

func newTestEventCreatedMessage(
	t *testing.T,
	speed float64,
	batteryLevel float64,
	lat float64,
	lon float64,
) messaging.EventCreatedMessage {
	t.Helper()

	eventID := uuid.New()
	vehicleID := uuid.New()
	deviceID := uuid.New()

	return messaging.EventCreatedMessage{
		EventID:      eventID.String(),
		VehicleID:    vehicleID.String(),
		DeviceID:     deviceID.String(),
		EventType:    "telemetry",
		EventTime:    "2026-05-11T10:15:00Z",
		Speed:        &speed,
		BatteryLevel: &batteryLevel,
		Lat:          &lat,
		Lon:          &lon,
	}
}

func newTestEventCreatedMessageWithoutLocation(
	t *testing.T,
	speed float64,
	batteryLevel float64,
) messaging.EventCreatedMessage {
	t.Helper()

	eventID := uuid.New()
	vehicleID := uuid.New()
	deviceID := uuid.New()

	return messaging.EventCreatedMessage{
		EventID:      eventID.String(),
		VehicleID:    vehicleID.String(),
		DeviceID:     deviceID.String(),
		EventType:    "telemetry",
		EventTime:    "2026-05-11T10:15:00Z",
		Speed:        &speed,
		BatteryLevel: &batteryLevel,
		Lat:          nil,
		Lon:          nil,
	}
}
