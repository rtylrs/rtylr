package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"time"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/eventbus"
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
	cacher.Of[domain.DashboardStats]().Invalidate(ctx, []string{fmt.Sprintf("crm:dashboard:%d", businessID) + "|domain=v2"})
}

type Service struct {
	repo    repository
	ownerOf BusinessOwnerResolver
}

type BusinessOwnerResolver interface {
	GetBusinessOwnerID(context.Context, int64) (int64, error)
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) WithBusinessOwnerResolver(resolver BusinessOwnerResolver) *Service {
	s.ownerOf = resolver
	return s
}

func (s *Service) CaptureContactFromEvent(ctx context.Context, event eventbus.Event, contact crmcontract.CustomerContactCaptured) error {
	if event.ID == 0 || event.BusinessID <= 0 || event.SubjectType == "" || event.SubjectID <= 0 {
		return fmt.Errorf("customer contact event requires positive event, business, and subject ids")
	}
	if s.ownerOf == nil {
		return fmt.Errorf("customer contact capture requires a business owner resolver")
	}
	ownerID, err := s.ownerOf.GetBusinessOwnerID(ctx, event.BusinessID)
	if err != nil {
		return err
	}
	if ownerID <= 0 {
		return fmt.Errorf("customer contact capture resolved an invalid business owner")
	}
	if err := s.repo.CaptureCustomerContact(ctx, event, ownerID, contact); err != nil {
		return err
	}
	invalidateCustomerScope(ctx, event.BusinessID)
	return nil
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
		key+"|domain=v2",
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
		cacheKeys.Customer(businessID, customerID)+"|domain=v2",
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

type repository interface {
	SearchSalesContacts(context.Context, int64, string) ([]domain.SalesContact, error)
	CaptureCustomerContact(ctx context.Context, event eventbus.Event, ownerID int64, contact crmcontract.CustomerContactCaptured) error
	ApplyLoyaltyDelta(ctx context.Context, input domain.LoyaltyTransaction, delta int, tierFor func(int) string, onCommitted func()) error
	BulkCreateCustomers(ctx context.Context, businessID, ownerID int64, items []domain.CustomerImportItem, dedupe bool) (int, int, error)
	CreateCampaign(ctx context.Context, input domain.CampaignInput) (int64, error)
	CreateCustomer(ctx context.Context, input domain.CustomerInput) (int64, error)
	CreateLoyaltyTier(ctx context.Context, input domain.LoyaltyTierInput) (int64, error)
	CreateSegment(ctx context.Context, input domain.SegmentInput) (int64, error)
	DeleteCustomer(ctx context.Context, businessID, customerID int64) error
	DeleteLoyaltyTier(ctx context.Context, businessID, tierID int64) error
	DeleteSegment(ctx context.Context, businessID, segmentID int64) error
	ExportCustomerData(ctx context.Context, businessID, customerID int64) (domain.CustomerExport, error)
	GetCampaignByID(ctx context.Context, businessID, campaignID int64) (domain.Campaign, error)
	GetCampaignStats(ctx context.Context, businessID, campaignID int64) (domain.CampaignStats, error)
	GetCustomerByID(ctx context.Context, businessID, customerID int64) (domain.Customer, error)
	GetCustomerTimeline(ctx context.Context, businessID, customerID int64) ([]domain.TimelineEntry, error)
	GetDashboardStats(ctx context.Context, businessID int64) (domain.DashboardStats, error)
	GetLoyaltyConfig(ctx context.Context, businessID int64) (domain.LoyaltyConfig, error)
	GetSegmentByID(ctx context.Context, businessID, segmentID int64) (domain.Segment, error)
	GetSegmentCustomers(ctx context.Context, businessID, segmentID int64, page, pageSize int) ([]domain.Customer, int, error)
	InsertTimelineEntry(ctx context.Context, entry domain.TimelineEntry) error
	ListCampaigns(ctx context.Context, businessID int64, status string) ([]domain.Campaign, error)
	ListCustomerTimelinePage(ctx context.Context, businessID, customerID int64, cursorCreatedAt *time.Time, cursorID int64, pageSize int) ([]domain.TimelineEntry, error)
	ListCustomers(ctx context.Context, filter domain.CustomerListFilter) ([]domain.Customer, int, error)
	ListLoyaltyTiers(ctx context.Context, businessID int64) ([]domain.LoyaltyTier, error)
	ListLoyaltyTiersPage(ctx context.Context, businessID int64, page, pageSize int) ([]domain.LoyaltyTier, int64, error)
	ListLoyaltyTransactions(ctx context.Context, businessID, customerID int64, page, pageSize int) ([]domain.LoyaltyTransaction, int, error)
	ListSegments(ctx context.Context, businessID int64) ([]domain.Segment, error)
	MergeCustomers(ctx context.Context, input domain.MergeCustomersInput) error
	RecordVisit(ctx context.Context, businessID, customerID int64, amount float64) error
	UpdateCampaign(ctx context.Context, input domain.CampaignUpdateInput) error
	UpdateCampaignStatus(ctx context.Context, businessID, campaignID int64, status string) error
	UpdateCustomer(ctx context.Context, input domain.CustomerUpdateInput) error
	UpdateCustomerWithLoyalty(ctx context.Context, input domain.CustomerUpdateInput) error
	UpdateLoyaltyTier(ctx context.Context, input domain.LoyaltyTierUpdateInput) error
	UpdateSegment(ctx context.Context, input domain.SegmentUpdateInput) error
	UpsertLoyaltyConfig(ctx context.Context, input domain.LoyaltyConfigInput) error
}
