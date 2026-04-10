package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/views/internal/domain"
	"github.com/voxire/onion/internal/modules/views/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID, workspaceID int64) ([]domain.View, error) {
	return s.repo.List(ctx, ownerID, workspaceID)
}

func (s *Service) Create(ctx context.Context, input domain.ViewInput) (int64, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, input domain.ViewUpdateInput) error {
	return s.repo.Update(ctx, input)
}

func (s *Service) Delete(ctx context.Context, id, ownerID int64) error {
	return s.repo.Delete(ctx, id, ownerID)
}
