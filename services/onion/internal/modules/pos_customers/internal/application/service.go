package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/pos_customers/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_customers/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) Search(ctx context.Context, ownerID int64, query string) ([]domain.Customer, error) {
	return s.repo.Search(ctx, ownerID, query)
}
