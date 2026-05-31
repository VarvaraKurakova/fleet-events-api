package domain

import (
	"time"

	"github.com/google/uuid"
)

type VehicleState struct {
	VehicleID    uuid.UUID
	DeviceID     uuid.UUID
	EventID      uuid.UUID
	Lat          *float64
	Lon          *float64
	Speed        *float64
	BatteryLevel *float64
	Ignition     *bool
	EventTime    time.Time
	ReceivedAt   time.Time
}
