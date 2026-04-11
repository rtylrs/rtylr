package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/support_customers/internal/domain"
	"github.com/voxire/onion/internal/modules/support_customers/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) List(ctx context.Context, businessID int64, page, pageSize int) ([]domain.Customer, int, error) {
	return s.repo.List(ctx, businessID, page, pageSize)
}

func (s *Service) GetByID(ctx context.Context, businessID, id int64) (*domain.Customer, error) {
	return s.repo.GetByID(ctx, businessID, id)
}
