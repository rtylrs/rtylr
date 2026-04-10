package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/pos_gift_cards/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_gift_cards/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.GiftCard, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.GiftCardInput) (int64, error) {
	return s.repo.Create(ctx, ownerID, input)
}

func (s *Service) GetByCode(ctx context.Context, ownerID int64, code string) (*domain.GiftCard, error) {
	return s.repo.GetByCode(ctx, ownerID, code)
}

func (s *Service) Redeem(ctx context.Context, ownerID int64, input domain.RedeemInput) (*domain.GiftCard, error) {
	return s.repo.Redeem(ctx, ownerID, input)
}
