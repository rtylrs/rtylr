package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/loyalty/internal/domain"
	"github.com/voxire/onion/internal/modules/loyalty/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	loyaltyScope = "loyalty"
	loyaltyTTL   = 5 * time.Minute
)

type repository interface {
	ListTransactions(ctx context.Context, businessID, customerID int64, page, pageSize int) ([]domain.LoyaltyTransaction, int, error)
	CreateTransaction(ctx context.Context, input domain.AwardInput, txType string) error
	CreateRedeemTransaction(ctx context.Context, input domain.RedeemInput) error
	GetCustomerPointsBalance(ctx context.Context, businessID, customerID int64) (int64, error)
	GetConfig(ctx context.Context, businessID int64) (domain.LoyaltyConfig, error)
	UpsertConfig(ctx context.Context, input domain.LoyaltyConfigInput) error
	ListMembers(ctx context.Context, businessID int64, limit, offset int) ([]domain.LoyaltyMember, error)
	ListRewards(ctx context.Context, businessID int64, activeOnly bool) ([]domain.LoyaltyReward, error)
	UpsertReward(ctx context.Context, rw domain.LoyaltyReward) (int64, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateLoyaltyScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(loyaltyScope, businessID))
}

type loyaltyTxResult struct {
	Items []domain.LoyaltyTransaction
	Total int
}

func (s *Service) ListTransactions(ctx context.Context, businessID, customerID int64, page, pageSize int) ([]domain.LoyaltyTransaction, int, error) {
	key := cacheKeys.Scoped(loyaltyScope, businessID, fmt.Sprintf("tx|c=%d|p=%d|ps=%d", customerID, page, pageSize))
	res, err := cacher.Of[loyaltyTxResult]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (loyaltyTxResult, error) {
			items, total, err := s.repo.ListTransactions(ctx, businessID, customerID, page, pageSize)
			return loyaltyTxResult{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, loyaltyTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) AwardPoints(ctx context.Context, input domain.AwardInput) error {
	if err := s.repo.CreateTransaction(ctx, input, "award"); err != nil {
		return err
	}
	invalidateLoyaltyScope(ctx, input.BusinessID)
	return nil
}

func (s *Service) RedeemPoints(ctx context.Context, input domain.RedeemInput) error {
	// A redemption must respect the configured minimum and can never drive
	// the member's ledger balance negative.
	if cfg, cfgErr := s.repo.GetConfig(ctx, input.BusinessID); cfgErr == nil {
		if cfg.MinRedemptionPoints > 0 && input.Points < cfg.MinRedemptionPoints {
			return fmt.Errorf("%w: minimum redemption is %d points", domain.ErrInvalidLoyaltyConfig, cfg.MinRedemptionPoints)
		}
	} else if !errors.Is(cfgErr, domain.ErrLoyaltyConfigNotFound) {
		return cfgErr
	}
	balance, balErr := s.repo.GetCustomerPointsBalance(ctx, input.BusinessID, input.CustomerID)
	if balErr != nil {
		return balErr
	}
	if balance < int64(input.Points) {
		return domain.ErrInsufficientPoints
	}
	if err := s.repo.CreateRedeemTransaction(ctx, input); err != nil {
		return err
	}
	invalidateLoyaltyScope(ctx, input.BusinessID)
	return nil
}

func (s *Service) GetConfig(ctx context.Context, businessID int64) (domain.LoyaltyConfig, error) {
	return cacher.Of[domain.LoyaltyConfig]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(loyaltyScope, businessID, "config"),
		func(ctx context.Context) (domain.LoyaltyConfig, error) {
			return s.repo.GetConfig(ctx, businessID)
		},
		cacher.WithExpiry(ctx, loyaltyTTL),
	)
}

func (s *Service) UpdateConfig(ctx context.Context, input domain.LoyaltyConfigInput) error {
	if msg := domain.ValidateLoyaltyConfigInput(input); msg != "" {
		return fmt.Errorf("%w: %s", domain.ErrInvalidLoyaltyConfig, msg)
	}
	if err := s.repo.UpsertConfig(ctx, input); err != nil {
		return err
	}
	invalidateLoyaltyScope(ctx, input.BusinessID)
	return nil
}

func (s *Service) ListMembers(ctx context.Context, businessID int64, limit, offset int) ([]domain.LoyaltyMember, error) {
	key := cacheKeys.Scoped(loyaltyScope, businessID, fmt.Sprintf("members|l=%d|o=%d", limit, offset))
	return cacher.Of[[]domain.LoyaltyMember]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.LoyaltyMember, error) {
			return s.repo.ListMembers(ctx, businessID, limit, offset)
		},
		cacher.WithExpiry(ctx, loyaltyTTL),
	)
}

func (s *Service) ListRewards(ctx context.Context, businessID int64, activeOnly bool) ([]domain.LoyaltyReward, error) {
	key := cacheKeys.Scoped(loyaltyScope, businessID, fmt.Sprintf("rewards|active=%v", activeOnly))
	return cacher.Of[[]domain.LoyaltyReward]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.LoyaltyReward, error) {
			return s.repo.ListRewards(ctx, businessID, activeOnly)
		},
		cacher.WithExpiry(ctx, loyaltyTTL),
	)
}

func (s *Service) UpsertReward(ctx context.Context, rw domain.LoyaltyReward) (int64, error) {
	id, err := s.repo.UpsertReward(ctx, rw)
	if err != nil {
		return 0, err
	}
	invalidateLoyaltyScope(ctx, rw.BusinessID)
	return id, nil
}
