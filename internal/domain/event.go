package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID           uuid.UUID
	DeviceID     uuid.UUID
	VehicleID    uuid.UUID
	EventType    string
	EventTime    time.Time
	ReceivedAt   time.Time
	Lat          *float64
	Lon          *float64
	Speed        *float64
	BatteryLevel *float64
	Ignition     *bool
	Payload      json.RawMessage
	CreatedAt    time.Time
}
