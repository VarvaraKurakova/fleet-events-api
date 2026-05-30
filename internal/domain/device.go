package domain

import (
	"time"

	"github.com/google/uuid"
)

type Device struct {
	ID         uuid.UUID
	VehicleID  uuid.UUID
	ExternalID string
	Model      *string
	Status     string
	LastSeenAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
