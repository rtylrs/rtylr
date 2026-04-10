package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/cycles/internal/domain"
	"github.com/voxire/onion/internal/modules/cycles/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID int64, workspaceID *int64) ([]domain.Cycle, error) {
	return s.repo.List(ctx, ownerID, workspaceID)
}

func (s *Service) Create(ctx context.Context, input domain.CycleInput) (int64, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, input domain.CycleUpdateInput) error {
	return s.repo.Update(ctx, input)
}
