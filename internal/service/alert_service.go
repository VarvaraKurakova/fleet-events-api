package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/VarvaraKurakova/fleet-events-api/internal/apperrors"
	"github.com/VarvaraKurakova/fleet-events-api/internal/domain"
	"github.com/VarvaraKurakova/fleet-events-api/internal/messaging"
)

const (
	speedLimitThreshold = 90.0
	lowBatteryThreshold = 0.15
)

type AlertRepository interface {
	Create(
		ctx context.Context,
		vehicleID uuid.UUID,
		deviceID uuid.UUID,
		eventID *uuid.UUID,
		alertType string,
		severity string,
		message string,
	) (domain.Alert, error)
	List(ctx context.Context, filter domain.AlertListFilter) ([]domain.Alert, error)
	Resolve(ctx context.Context, id uuid.UUID) (domain.Alert, error)
}

type AlertListFilter struct {
	Status    string
	Type      string
	VehicleID *uuid.UUID
	Limit     int
	Offset    int
}

type AlertService struct {
	alertRepository AlertRepository
}

func NewAlertService(alertRepository AlertRepository) *AlertService {
	return &AlertService{
		alertRepository: alertRepository,
	}
}

func (s *AlertService) ProcessEvent(
	ctx context.Context,
	message messaging.EventCreatedMessage,
) ([]domain.Alert, error) {
	eventID, err := uuid.Parse(message.EventID)
	if err != nil {
		return nil, apperrors.ErrInvalidInput
	}

	vehicleID, err := uuid.Parse(message.VehicleID)
	if err != nil {
		return nil, apperrors.ErrInvalidInput
	}

	deviceID, err := uuid.Parse(message.DeviceID)
	if err != nil {
		return nil, apperrors.ErrInvalidInput
	}

	alerts := make([]domain.Alert, 0)

	if message.Speed != nil && *message.Speed > speedLimitThreshold {
		alert, err := s.alertRepository.Create(
			ctx,
			vehicleID,
			deviceID,
			&eventID,
			domain.AlertTypeSpeedLimitExceeded,
			domain.AlertSeverityWarning,
			fmt.Sprintf("vehicle speed %.1f is above limit %.1f", *message.Speed, speedLimitThreshold),
		)
		if err != nil {
			return nil, err
		}

		alerts = append(alerts, alert)
	}

	if message.BatteryLevel != nil && *message.BatteryLevel < lowBatteryThreshold {
		alert, err := s.alertRepository.Create(
			ctx,
			vehicleID,
			deviceID,
			&eventID,
			domain.AlertTypeLowBattery,
			domain.AlertSeverityWarning,
			fmt.Sprintf("device battery level %.2f is below threshold %.2f", *message.BatteryLevel, lowBatteryThreshold),
		)
		if err != nil {
			return nil, err
		}

		alerts = append(alerts, alert)
	}

	if isInvalidLocation(message.Lat, message.Lon) {
		alert, err := s.alertRepository.Create(
			ctx,
			vehicleID,
			deviceID,
			&eventID,
			domain.AlertTypeInvalidLocation,
			domain.AlertSeverityWarning,
			"event contains invalid location",
		)
		if err != nil {
			return nil, err
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (s *AlertService) List(ctx context.Context, filter domain.AlertListFilter) ([]domain.Alert, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	if filter.Limit > 100 {
		filter.Limit = 100
	}

	if filter.Offset < 0 {
		filter.Offset = 0
	}

	return s.alertRepository.List(ctx, filter)
}

func (s *AlertService) Resolve(ctx context.Context, id uuid.UUID) (domain.Alert, error) {
	if id == uuid.Nil {
		return domain.Alert{}, apperrors.ErrInvalidInput
	}

	return s.alertRepository.Resolve(ctx, id)
}

func isInvalidLocation(lat *float64, lon *float64) bool {
	if lat == nil || lon == nil {
		return true
	}

	if *lat < -90 || *lat > 90 {
		return true
	}

	if *lon < -180 || *lon > 180 {
		return true
	}

	return false
}
