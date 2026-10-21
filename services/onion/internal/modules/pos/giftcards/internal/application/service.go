package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/pos/giftcards/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	posGiftCardsScope = "pos_gift_cards"
	posGiftCardsTTL   = 1 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.GiftCard, error)
	CreateForOwner(ctx context.Context, businessID, ownerID int64, input domain.GiftCardInput) (int64, error)
	GetByCode(ctx context.Context, businessID int64, code string) (*domain.GiftCard, error)
	RedeemForOrder(ctx context.Context, businessID, actorID int64, input domain.RedeemInput) (*domain.GiftCard, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func invalidatePosGiftCardsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(posGiftCardsScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.GiftCard, error) {
	return cacher.Of[[]domain.GiftCard]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(posGiftCardsScope, businessID, "list|domain=v2"),
		func(ctx context.Context) ([]domain.GiftCard, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, posGiftCardsTTL),
	)
}

func (s *Service) CreateForOwner(ctx context.Context, businessID, ownerID int64, input domain.GiftCardInput) (int64, error) {
	if err := domain.ValidateGiftCardInput(input); err != nil {
		return 0, err
	}
	id, err := s.repo.CreateForOwner(ctx, businessID, ownerID, input)
	if err != nil {
		return 0, err
	}
	invalidatePosGiftCardsScope(ctx, businessID)
	return id, nil
}

func (s *Service) GetByCode(ctx context.Context, businessID int64, code string) (*domain.GiftCard, error) {
	key := cacheKeys.Scoped(posGiftCardsScope, businessID, fmt.Sprintf("code=%s", code))
	return cacher.Of[*domain.GiftCard]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (*domain.GiftCard, error) {
			return s.repo.GetByCode(ctx, businessID, code)
		},
		cacher.WithExpiry(ctx, posGiftCardsTTL),
	)
}

func (s *Service) RedeemForOrder(ctx context.Context, businessID, actorID int64, input domain.RedeemInput) (*domain.GiftCard, error) {
	if err := domain.ValidateRedeemInput(input); err != nil {
		return nil, err
	}
	card, err := s.repo.RedeemForOrder(ctx, businessID, actorID, input)
	if err != nil {
		return nil, err
	}
	invalidatePosGiftCardsScope(ctx, businessID)
	return card, nil
}
