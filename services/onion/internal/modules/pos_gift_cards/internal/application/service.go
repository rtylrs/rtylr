package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/pos_gift_cards/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_gift_cards/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	posGiftCardsScope = "pos_gift_cards"
	posGiftCardsTTL   = 1 * time.Minute
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.GiftCard, error)
	Create(ctx context.Context, ownerID int64, input domain.GiftCardInput) (int64, error)
	GetByCode(ctx context.Context, ownerID int64, code string) (*domain.GiftCard, error)
	Redeem(ctx context.Context, ownerID int64, input domain.RedeemInput) (*domain.GiftCard, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidatePosGiftCardsScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(posGiftCardsScope, ownerID))
	}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.GiftCard, error) {
	return cacher.Of[[]domain.GiftCard]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(posGiftCardsScope, ownerID, "list"),
		func(ctx context.Context) ([]domain.GiftCard, error) {
			return s.repo.List(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, posGiftCardsTTL),
	)
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.GiftCardInput) (int64, error) {
	if err := domain.ValidateGiftCardInput(input); err != nil {
		return 0, err
	}
	id, err := s.repo.Create(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}
	invalidatePosGiftCardsScope(ctx, ownerID)
	return id, nil
}

func (s *Service) GetByCode(ctx context.Context, ownerID int64, code string) (*domain.GiftCard, error) {
	key := cacheKeys.Scoped(posGiftCardsScope, ownerID, fmt.Sprintf("code=%s", code))
	return cacher.Of[*domain.GiftCard]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (*domain.GiftCard, error) {
			return s.repo.GetByCode(ctx, ownerID, code)
		},
		cacher.WithExpiry(ctx, posGiftCardsTTL),
	)
}

func (s *Service) Redeem(ctx context.Context, ownerID int64, input domain.RedeemInput) (*domain.GiftCard, error) {
	if err := domain.ValidateRedeemInput(input); err != nil {
		return nil, err
	}
	card, err := s.repo.Redeem(ctx, ownerID, input)
	if err != nil {
		return nil, err
	}
	invalidatePosGiftCardsScope(ctx, ownerID)
	return card, nil
}
