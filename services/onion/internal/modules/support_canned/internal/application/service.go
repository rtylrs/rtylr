package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/support_canned/internal/domain"
	"github.com/voxire/onion/internal/modules/support_canned/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.CannedResponse, error) {
	return s.repo.List(ctx, businessID)
}

func (s *Service) Create(ctx context.Context, input domain.CannedResponseInput) (int64, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.CannedResponseInput) error {
	return s.repo.Update(ctx, businessID, id, input)
}

func (s *Service) Delete(ctx context.Context, businessID, id int64) error {
	return s.repo.Delete(ctx, businessID, id)
}
