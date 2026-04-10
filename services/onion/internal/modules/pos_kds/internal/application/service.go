package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/pos_kds/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_kds/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) ListActiveOrders(ctx context.Context, ownerID int64) ([]domain.KDSOrder, error) {
	return s.repo.ListActiveOrders(ctx, ownerID)
}

func (s *Service) UpdateItemStatus(ctx context.Context, input domain.UpdateItemStatusInput) error {
	return s.repo.UpdateItemStatus(ctx, input)
}
