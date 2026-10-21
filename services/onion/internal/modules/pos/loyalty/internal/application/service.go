package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/pos/loyalty/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	posLoyaltyScope = "pos_loyalty"
	posLoyaltyTTL   = 2 * time.Minute
)

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func invalidatePosLoyaltyScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(posLoyaltyScope, businessID))
}

func (s *Service) GetConfig(ctx context.Context, businessID int64) (domain.LoyaltyConfig, error) {
	return cacher.Of[domain.LoyaltyConfig]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(posLoyaltyScope, businessID, "config|domain=v2"),
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
	return cacher.Of[[]domain.LoyaltyTransaction]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) ([]domain.LoyaltyTransaction, error) {
			return s.repo.ListTransactions(ctx, businessID, customerID)
		},
		cacher.WithExpiry(ctx, posLoyaltyTTL),
	)
}

func (s *Service) GetBalance(ctx context.Context, businessID, customerID int64) (int32, error) {
	key := cacheKeys.Scoped(posLoyaltyScope, businessID, fmt.Sprintf("bal|c=%d", customerID))
	return cacher.Of[int32]().GetOrComputeSingle(ctx, key+"|domain=v2",
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
	if input.OrderID == nil || *input.OrderID == 0 {
		return 0, domain.ErrOrderRequired
	}
	cfg, err := s.repo.GetConfig(ctx, input.BusinessID)
	if err != nil {
		return 0, err
	}
	if !cfg.IsEnabled {
		return 0, domain.ErrLoyaltyNotEnabled
	}
	orderTotal, err := s.repo.GetPaidOrderTotal(ctx, input.BusinessID, input.CustomerID, *input.OrderID)
	if err != nil {
		return 0, err
	}
	points := int32(orderTotal * cfg.PointsPerDollar)
	if points <= 0 {
		return 0, domain.ErrOrderNotPaid
	}
	note := "Points earned from paid order"
	id, err := s.repo.CreateBalanceTransaction(ctx, input.BusinessID, input.CustomerID, input.OrderID, "earn", points, note)
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
	cfg, err := s.repo.GetConfig(ctx, input.BusinessID)
	if err != nil {
		return 0, err
	}
	if !cfg.IsEnabled {
		return 0, domain.ErrLoyaltyNotEnabled
	}
	if input.Points < cfg.MinRedeemPoints {
		return 0, domain.ErrMinRedeemNotReached
	}
	note := fmt.Sprintf("Points redeemed at rate %.4f", cfg.RedemptionRate)
	id, err := s.repo.CreateBalanceTransaction(ctx, input.BusinessID, input.CustomerID, nil, "redeem", -input.Points, note)
	if err != nil {
		return 0, err
	}
	invalidatePosLoyaltyScope(ctx, input.BusinessID)
	return id, nil
}

type repository interface {
	CreateBalanceTransaction(ctx context.Context, businessID, customerID int64, orderID *int64, txnType string, points int32, note string) (int64, error)
	GetBalance(ctx context.Context, businessID, customerID int64) (int32, error)
	GetConfig(ctx context.Context, businessID int64) (domain.LoyaltyConfig, error)
	GetPaidOrderTotal(ctx context.Context, businessID, customerID, orderID int64) (float64, error)
	ListTransactions(ctx context.Context, businessID, customerID int64) ([]domain.LoyaltyTransaction, error)
	UpsertConfig(ctx context.Context, input domain.UpdateLoyaltyConfigInput) error
}
