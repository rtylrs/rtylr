package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/pos_locations/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_locations/internal/infrastructure"
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.Location, error)
	Create(ctx context.Context, input domain.LocationInput) (int64, error)
	Update(ctx context.Context, input domain.LocationUpdateInput) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Location, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *Service) Create(ctx context.Context, input domain.LocationInput) (int64, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, input domain.LocationUpdateInput) error {
	return s.repo.Update(ctx, input)
}
