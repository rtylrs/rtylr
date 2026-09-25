package application

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math"
	"time"

	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/customers/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	dashboardCacheTTL = 5 * time.Minute
	loyaltyCacheTTL   = 30 * time.Minute
	customerCacheTTL  = 2 * time.Minute
)

// invalidateCustomerScope wipes every cache entry derived from a customer for
// the business. Customer CRUD and loyalty mutations feed both the customer
// list/detail cache and the dashboard aggregate, so the two derived families
// must be invalidated together.
func invalidateCustomerScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.CustomerPrefix(businessID))
	cacher.Of[domain.DashboardStats]().Invalidate(ctx, []string{fmt.Sprintf("crm:dashboard:%d", businessID)})
}

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

// ─── Customers ────────────────────────────────────────────────────────────────

func (s *Service) ListCustomers(ctx context.Context, filter domain.CustomerListFilter) (domain.CustomerListResult, error) {
	// Filter values are arbitrary user input; hashing them keeps separator
	// characters like '|' from colliding two different filter sets onto one
	// cached page within the same business.
	filterKey := ""
	if filter.Search != "" || filter.Tier != "" || filter.Tag != "" {
		sum := sha256.Sum256([]byte(fmt.Sprintf("s=%s\nt=%s\ntag=%s", filter.Search, filter.Tier, filter.Tag)))
		filterKey = hex.EncodeToString(sum[:8])
	}
	key := fmt.Sprintf(
		"%s|f=%s|p=%d|ps=%d",
		cacheKeys.CustomersList(filter.BusinessID),
		filterKey, filter.Page, filter.PageSize,
	)
	return cacher.Of[domain.CustomerListResult]().GetOrComputeSingle(
		ctx,
		key,
		func(ctx context.Context) (domain.CustomerListResult, error) {
			customers, total, err := s.repo.ListCustomers(ctx, filter)
			if err != nil {
				return domain.CustomerListResult{}, err
			}
			return domain.CustomerListResult{
				Customers: customers,
				Total:     total,
				Page:      filter.Page,
				PageSize:  filter.PageSize,
			}, nil
		},
		cacher.WithExpiry(ctx, customerCacheTTL),
	)
}

func (s *Service) GetCustomer(ctx context.Context, businessID, customerID int64) (domain.Customer, error) {
	return cacher.Of[domain.Customer]().GetOrComputeSingle(
		ctx,
		cacheKeys.Customer(businessID, customerID),
		func(ctx context.Context) (domain.Customer, error) {
			return s.repo.GetCustomerByID(ctx, businessID, customerID)
		},
		cacher.WithExpiry(ctx, customerCacheTTL),
	)
}

func (s *Service) CreateCustomer(ctx context.Context, input domain.CustomerInput) (int64, error) {
	id, err := s.repo.CreateCustomer(ctx, input)
	if err == nil {
		invalidateCustomerScope(ctx, input.BusinessID)
	}
	return id, err
}

func (s *Service) UpdateCustomer(ctx context.Context, input domain.CustomerUpdateInput) error {
	if input.LoyaltyTier != "" {
		if err := s.repo.UpdateCustomerWithLoyalty(ctx, input); err != nil {
			return err
		}
	} else if err := s.repo.UpdateCustomer(ctx, input); err != nil {
		return err
	}
	invalidateCustomerScope(ctx, input.BusinessID)
	return nil
}

func (s *Service) DeleteCustomer(ctx context.Context, businessID, customerID int64) error {
	if err := s.repo.DeleteCustomer(ctx, businessID, customerID); err != nil {
		return err
	}
	invalidateCustomerScope(ctx, businessID)
	return nil
}

func (s *Service) RecordVisit(ctx context.Context, input domain.RecordVisitInput) error {
	if err := s.repo.RecordVisit(ctx, input.BusinessID, input.CustomerID, input.Amount); err != nil {
		return err
	}

	// The visit write feeds the dashboard counters, so the cache must be
	// busted on every path out of this function — including the award
	// block's early return when the loyalty config lookup fails.
	defer s.invalidateDashboardCache(ctx, input.BusinessID)

	s.insertTimeline(ctx, input.BusinessID, input.OwnerID, input.CustomerID,
		"visit", fmt.Sprintf("Visit recorded (spend: %.2f)", input.Amount),
		map[string]any{"amount": input.Amount})

	if input.AwardPoints && input.Amount > 0 {
		cfg, err := s.GetLoyaltyConfig(ctx, input.BusinessID)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to get loyalty config for visit award", logger.Error(err))

			return nil
		}

		if cfg.IsEnabled && cfg.PointsPerCurrencyUnit > 0 {
			points := loyaltyPointsFromCents(input.Amount, cfg.PointsPerCurrencyUnit)
			if points > 0 {
				if awardErr := s.AwardPoints(ctx, domain.AwardPointsInput{
					OwnerID:     input.OwnerID,
					BusinessID:  input.BusinessID,
					CustomerID:  input.CustomerID,
					Points:      points,
					Description: fmt.Sprintf("Points earned for visit (spend: %.2f)", input.Amount),
				}); awardErr != nil {
					logger.Ctx(ctx).Error("Failed to auto-award points on visit", logger.Error(awardErr))
				}
			}
		}
	}

	return nil
}

// loyaltyPointsFromCents rounds the purchase to integer minor units before
// applying the configured rate. This avoids binary-float truncation such as
// 99.99 * 0.1 becoming 9 points instead of the configured 10-point result.
func loyaltyPointsFromCents(amount, pointsPerCurrencyUnit float64) int {
	amountCents := math.Round(amount * 100)
	return int(math.Round(amountCents * pointsPerCurrencyUnit / 100))
}

// ─── Loyalty ──────────────────────────────────────────────────────────────────

func (s *Service) AwardPoints(ctx context.Context, input domain.AwardPointsInput) error {
	if err := domain.ValidateLoyaltyPoints(int64(input.Points)); err != nil {
		return err
	}

	cfg, cfgErr := s.GetLoyaltyConfig(ctx, input.BusinessID)
	if cfgErr != nil {
		logger.Ctx(ctx).Error("Failed to get loyalty config, using default tier", logger.Error(cfgErr))
	}
	return s.repo.ApplyLoyaltyDelta(ctx, domain.LoyaltyTransaction{
		OwnerID:       input.OwnerID,
		BusinessID:    input.BusinessID,
		CustomerID:    input.CustomerID,
		Type:          "earn",
		Points:        input.Points,
		Description:   input.Description,
		ReferenceID:   input.ReferenceID,
		ReferenceType: input.ReferenceType,
	}, input.Points, func(balance int) string { return RecalculateTier(balance, cfg) }, func() {
		s.insertTimeline(ctx, input.BusinessID, input.OwnerID, input.CustomerID,
			"award", input.Description, map[string]any{"points": input.Points})
		invalidateCustomerScope(ctx, input.BusinessID)
	})
}

func (s *Service) RedeemPoints(ctx context.Context, input domain.RedeemPointsInput) error {
	if err := domain.ValidateLoyaltyPoints(int64(input.Points)); err != nil {
		return err
	}

	cfg, cfgErr := s.GetLoyaltyConfig(ctx, input.BusinessID)
	if cfgErr != nil {
		logger.Ctx(ctx).Error("Failed to get loyalty config, using default tier", logger.Error(cfgErr))
	}
	return s.repo.ApplyLoyaltyDelta(ctx, domain.LoyaltyTransaction{
		OwnerID:       input.OwnerID,
		BusinessID:    input.BusinessID,
		CustomerID:    input.CustomerID,
		Type:          "redeem",
		Points:        input.Points,
		Description:   input.Description,
		ReferenceID:   input.ReferenceID,
		ReferenceType: input.ReferenceType,
	}, -input.Points, func(balance int) string { return RecalculateTier(balance, cfg) }, func() {
		s.insertTimeline(ctx, input.BusinessID, input.OwnerID, input.CustomerID,
			"reward", input.Description, map[string]any{"points": input.Points})
		invalidateCustomerScope(ctx, input.BusinessID)
	})
}
