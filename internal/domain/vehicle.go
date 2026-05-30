package domain

import (
	"time"

	"github.com/google/uuid"
)

type Vehicle struct {
	ID          uuid.UUID
	FleetID     uuid.UUID
	PlateNumber string
	VIN         *string
	Type        string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
