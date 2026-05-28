package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/pos_loyalty/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_loyalty/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	posLoyaltyScope = "pos_loyalty"
	posLoyaltyTTL   = 2 * time.Minute
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidatePosLoyaltyScope(ctx context.Context, businessID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(posLoyaltyScope, businessID))
	}
}

func (s *Service) GetConfig(ctx context.Context, businessID int64) (domain.LoyaltyConfig, error) {
	return cacher.Of[domain.LoyaltyConfig]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(posLoyaltyScope, businessID, "config"),
		func(ctx context.Context) (domain.LoyaltyConfig, error) {
			return s.repo.GetConfig(ctx, businessID)
		},
		cacher.WithExpiry(ctx, posLoyaltyTTL),
	)
}

func (s *Service) UpdateConfig(ctx context.Context, input domain.UpdateLoyaltyConfigInput) error {
	if err := s.repo.UpsertConfig(ctx, input); err != nil {
		return err
	}
	invalidatePosLoyaltyScope(ctx, input.BusinessID)
	return nil
}

func (s *Service) ListTransactions(ctx context.Context, businessID, customerID int64) ([]domain.LoyaltyTransaction, error) {
	key := cacheKeys.Scoped(posLoyaltyScope, businessID, fmt.Sprintf("tx|c=%d", customerID))
	return cacher.Of[[]domain.LoyaltyTransaction]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.LoyaltyTransaction, error) {
			return s.repo.ListTransactions(ctx, businessID, customerID)
		},
		cacher.WithExpiry(ctx, posLoyaltyTTL),
	)
}

func (s *Service) GetBalance(ctx context.Context, businessID, customerID int64) (int32, error) {
	key := cacheKeys.Scoped(posLoyaltyScope, businessID, fmt.Sprintf("bal|c=%d", customerID))
	return cacher.Of[int32]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (int32, error) {
			return s.repo.GetBalance(ctx, businessID, customerID)
		},
		cacher.WithExpiry(ctx, posLoyaltyTTL),
	)
}

func (s *Service) EarnPoints(ctx context.Context, input domain.EarnPointsInput) (int64, error) {
	if err := domain.ValidateEarnPointsInput(input); err != nil {
		return 0, err
	}
	currentBalance, err := s.repo.GetBalance(ctx, input.BusinessID, input.CustomerID)
	if err != nil {
		return 0, err
	}
	newBalance := currentBalance + input.Points
	note := "Points earned"
	id, err := s.repo.CreateTransaction(ctx, input.BusinessID, input.CustomerID, input.OrderID, "earn", input.Points, newBalance, &note)
	if err != nil {
		return 0, err
	}
	invalidatePosLoyaltyScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) RedeemPoints(ctx context.Context, input domain.RedeemPointsInput) (int64, error) {
	if err := domain.ValidateRedeemPointsInput(input); err != nil {
		return 0, err
	}
	currentBalance, err := s.repo.GetBalance(ctx, input.BusinessID, input.CustomerID)
	if err != nil {
		return 0, err
	}
	if currentBalance < input.Points {
		return 0, domain.ErrInsufficientPoints
	}
	newBalance := currentBalance - input.Points
	note := "Points redeemed"
	id, err := s.repo.CreateTransaction(ctx, input.BusinessID, input.CustomerID, nil, "redeem", -input.Points, newBalance, &note)
	if err != nil {
		return 0, err
	}
	invalidatePosLoyaltyScope(ctx, input.BusinessID)
	return id, nil
}
