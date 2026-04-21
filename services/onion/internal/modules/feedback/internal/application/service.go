package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/feedback/internal/domain"
	"github.com/voxire/onion/internal/modules/feedback/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) List(ctx context.Context, ownerID int64, filters domain.ListFilters, page, pageSize int) ([]domain.Feedback, int, error) {
	return s.repo.List(ctx, ownerID, filters, page, pageSize)
}

func (s *Service) Create(ctx context.Context, input domain.FeedbackInput) (int64, error) {
	return s.repo.Create(ctx, input)
}
