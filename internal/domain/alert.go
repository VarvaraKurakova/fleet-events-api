package domain

import (
	"time"

	"github.com/google/uuid"
)

type Alert struct {
	ID         uuid.UUID
	VehicleID  uuid.UUID
	DeviceID   uuid.UUID
	EventID    *uuid.UUID
	Type       string
	Severity   string
	Message    string
	Status     string
	CreatedAt  time.Time
	ResolvedAt *time.Time
}

const (
	AlertTypeSpeedLimitExceeded = "speed_limit_exceeded"
	AlertTypeLowBattery         = "low_battery"
	AlertTypeInvalidLocation    = "invalid_location"

	AlertSeverityWarning = "warning"

	AlertStatusOpen     = "open"
	AlertStatusResolved = "resolved"
)
