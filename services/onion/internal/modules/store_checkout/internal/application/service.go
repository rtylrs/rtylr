package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/store_checkout/internal/domain"
	"github.com/voxire/onion/internal/modules/store_checkout/internal/infrastructure"
)

type repository interface {
	Create(ctx context.Context, ownerID int64, input domain.CheckoutCreateInput) (int64, error)
	Complete(ctx context.Context, ownerID, id int64) error
}

type Service struct{ repo repository }

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.CheckoutCreateInput) (int64, error) {
	return s.repo.Create(ctx, ownerID, input)
}

func (s *Service) Complete(ctx context.Context, ownerID, id int64) error {
	return s.repo.Complete(ctx, ownerID, id)
}
