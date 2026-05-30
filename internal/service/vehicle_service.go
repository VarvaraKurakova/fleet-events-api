package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/VarvaraKurakova/fleet-events-api/internal/apperrors"
	"github.com/VarvaraKurakova/fleet-events-api/internal/domain"
)

type VehicleRepository interface {
	Create(
		ctx context.Context,
		fleetID uuid.UUID,
		plateNumber string,
		vin *string,
		vehicleType string,
	) (domain.Vehicle, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Vehicle, error)
	List(ctx context.Context) ([]domain.Vehicle, error)
}

type VehicleService struct {
	vehicleRepository VehicleRepository
	fleetRepository   FleetRepository
}

func NewVehicleService(
	vehicleRepository VehicleRepository,
	fleetRepository FleetRepository,
) *VehicleService {
	return &VehicleService{
		vehicleRepository: vehicleRepository,
		fleetRepository:   fleetRepository,
	}
}

func (s *VehicleService) Create(
	ctx context.Context,
	fleetID uuid.UUID,
	plateNumber string,
	vin *string,
	vehicleType string,
) (domain.Vehicle, error) {
	plateNumber = strings.TrimSpace(plateNumber)
	vehicleType = strings.TrimSpace(vehicleType)

	if vin != nil {
		trimmedVIN := strings.TrimSpace(*vin)
		if trimmedVIN == "" {
			vin = nil
		} else {
			vin = &trimmedVIN
		}
	}

	if fleetID == uuid.Nil || plateNumber == "" || vehicleType == "" {
		return domain.Vehicle{}, apperrors.ErrInvalidInput
	}

	if _, err := s.fleetRepository.GetByID(ctx, fleetID); err != nil {
		return domain.Vehicle{}, err
	}

	return s.vehicleRepository.Create(ctx, fleetID, plateNumber, vin, vehicleType)
}

func (s *VehicleService) GetByID(ctx context.Context, id uuid.UUID) (domain.Vehicle, error) {
	return s.vehicleRepository.GetByID(ctx, id)
}

func (s *VehicleService) List(ctx context.Context) ([]domain.Vehicle, error) {
	return s.vehicleRepository.List(ctx)
}
