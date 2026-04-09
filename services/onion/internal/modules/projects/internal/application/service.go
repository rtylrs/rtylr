package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/projects/internal/domain"
	"github.com/voxire/onion/internal/modules/projects/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Project, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *Service) Create(ctx context.Context, input domain.ProjectInput) (int64, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, input domain.ProjectUpdateInput) error {
	return s.repo.Update(ctx, input)
}

func (s *Service) Delete(ctx context.Context, ownerID, id int64) error {
	return s.repo.Delete(ctx, ownerID, id)
}
