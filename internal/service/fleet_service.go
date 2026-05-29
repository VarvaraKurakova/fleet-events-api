package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/VarvaraKurakova/fleet-events-api/internal/apperrors"
	"github.com/VarvaraKurakova/fleet-events-api/internal/domain"
)

type FleetRepository interface {
	Create(ctx context.Context, name string) (domain.Fleet, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Fleet, error)
	List(ctx context.Context) ([]domain.Fleet, error)
}

type FleetService struct {
	repository FleetRepository
}

func NewFleetService(repository FleetRepository) *FleetService {
	return &FleetService{
		repository: repository,
	}
}

func (s *FleetService) Create(ctx context.Context, name string) (domain.Fleet, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Fleet{}, apperrors.ErrInvalidInput
	}

	return s.repository.Create(ctx, name)
}

func (s *FleetService) GetByID(ctx context.Context, id uuid.UUID) (domain.Fleet, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *FleetService) List(ctx context.Context) ([]domain.Fleet, error) {
	return s.repository.List(ctx)
}
