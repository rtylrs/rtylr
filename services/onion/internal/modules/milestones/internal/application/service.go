package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/milestones/internal/domain"
	"github.com/voxire/onion/internal/modules/milestones/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID int64, projectID *int64) ([]domain.Milestone, error) {
	return s.repo.List(ctx, ownerID, projectID)
}

func (s *Service) Create(ctx context.Context, input domain.MilestoneInput) (int64, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, input domain.MilestoneUpdateInput) error {
	return s.repo.Update(ctx, input)
}
