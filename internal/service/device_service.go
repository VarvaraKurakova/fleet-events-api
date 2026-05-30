package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/VarvaraKurakova/fleet-events-api/internal/apperrors"
	"github.com/VarvaraKurakova/fleet-events-api/internal/domain"
)

type DeviceRepository interface {
	Create(
		ctx context.Context,
		vehicleID uuid.UUID,
		externalID string,
		model *string,
	) (domain.Device, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Device, error)
	GetByExternalID(ctx context.Context, externalID string) (domain.Device, error)
	List(ctx context.Context) ([]domain.Device, error)
	UpdateLastSeen(ctx context.Context, id uuid.UUID, lastSeenAt time.Time) error
}

type DeviceService struct {
	deviceRepository  DeviceRepository
	vehicleRepository VehicleRepository
}

func NewDeviceService(
	deviceRepository DeviceRepository,
	vehicleRepository VehicleRepository,
) *DeviceService {
	return &DeviceService{
		deviceRepository:  deviceRepository,
		vehicleRepository: vehicleRepository,
	}
}

func (s *DeviceService) Create(
	ctx context.Context,
	vehicleID uuid.UUID,
	externalID string,
	model *string,
) (domain.Device, error) {
	externalID = strings.TrimSpace(externalID)

	if model != nil {
		trimmedModel := strings.TrimSpace(*model)
		if trimmedModel == "" {
			model = nil
		} else {
			model = &trimmedModel
		}
	}

	if vehicleID == uuid.Nil || externalID == "" {
		return domain.Device{}, apperrors.ErrInvalidInput
	}

	if _, err := s.vehicleRepository.GetByID(ctx, vehicleID); err != nil {
		return domain.Device{}, err
	}

	return s.deviceRepository.Create(ctx, vehicleID, externalID, model)
}

func (s *DeviceService) GetByID(ctx context.Context, id uuid.UUID) (domain.Device, error) {
	return s.deviceRepository.GetByID(ctx, id)
}

func (s *DeviceService) GetByExternalID(ctx context.Context, externalID string) (domain.Device, error) {
	externalID = strings.TrimSpace(externalID)
	if externalID == "" {
		return domain.Device{}, apperrors.ErrInvalidInput
	}

	return s.deviceRepository.GetByExternalID(ctx, externalID)
}

func (s *DeviceService) List(ctx context.Context) ([]domain.Device, error) {
	return s.deviceRepository.List(ctx)
}

func (s *DeviceService) UpdateLastSeen(ctx context.Context, id uuid.UUID, lastSeenAt time.Time) error {
	if id == uuid.Nil {
		return apperrors.ErrInvalidInput
	}

	return s.deviceRepository.UpdateLastSeen(ctx, id, lastSeenAt)
}
