package application

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
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

// invalidateCustomerScope wipes every customer-related cache entry for the
// business. Called by every mutation (create/update/delete/loyalty change).
func invalidateCustomerScope(ctx context.Context, businessID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.CustomerPrefix(businessID))
	}
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

	s.invalidateDashboardCache(ctx, input.BusinessID)

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
		s.invalidateDashboardCache(ctx, input.BusinessID)
	})
}

func (s *Service) RedeemPoints(ctx context.Context, input domain.RedeemPointsInput) error {
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
		s.invalidateDashboardCache(ctx, input.BusinessID)
	})
}

func (s *Service) ListLoyaltyTransactions(ctx context.Context, businessID, customerID int64, page, pageSize int) ([]domain.LoyaltyTransaction, int, error) {
	return s.repo.ListLoyaltyTransactions(ctx, businessID, customerID, page, pageSize)
}

func (s *Service) GetLoyaltyConfig(ctx context.Context, businessID int64) (domain.LoyaltyConfig, error) {
	cacheKey := fmt.Sprintf("crm:loyalty_config:%d", businessID)

	return cacher.Of[domain.LoyaltyConfig]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) (domain.LoyaltyConfig, error) {
		return s.repo.GetLoyaltyConfig(ctx, businessID)
	}, cacher.WithExpiry(ctx, loyaltyCacheTTL))
}

func (s *Service) UpsertLoyaltyConfig(ctx context.Context, input domain.LoyaltyConfigInput) error {
	if err := s.repo.UpsertLoyaltyConfig(ctx, input); err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("crm:loyalty_config:%d", input.BusinessID)
	cacher.Of[domain.LoyaltyConfig]().Invalidate(ctx, []string{cacheKey})

	return nil
}

// RecalculateTier returns the appropriate loyalty tier based on point balance and config thresholds.
func RecalculateTier(points int, cfg domain.LoyaltyConfig) string {
	platinum := cfg.PlatinumMin
	gold := cfg.GoldMin
	silver := cfg.SilverMin

	if platinum == 0 {
		platinum = 5000
	}

	if gold == 0 {
		gold = 1500
	}

	if silver == 0 {
		silver = 500
	}

	switch {
	case points >= platinum:
		return domain.TierPlatinum
	case points >= gold:
		return domain.TierGold
	case points >= silver:
		return domain.TierSilver
	default:
		return domain.TierBronze
	}
}

// ─── Segments ─────────────────────────────────────────────────────────────────

func (s *Service) ListSegments(ctx context.Context, businessID int64) ([]domain.Segment, error) {
	return s.repo.ListSegments(ctx, businessID)
}

func (s *Service) GetSegment(ctx context.Context, businessID, segmentID int64) (domain.Segment, error) {
	return s.repo.GetSegmentByID(ctx, businessID, segmentID)
}

func (s *Service) CreateSegment(ctx context.Context, input domain.SegmentInput) (int64, error) {
	return s.repo.CreateSegment(ctx, input)
}

func (s *Service) UpdateSegment(ctx context.Context, input domain.SegmentUpdateInput) error {
	return s.repo.UpdateSegment(ctx, input)
}

func (s *Service) DeleteSegment(ctx context.Context, businessID, segmentID int64) error {
	return s.repo.DeleteSegment(ctx, businessID, segmentID)
}

// ─── Campaigns ────────────────────────────────────────────────────────────────

func (s *Service) ListCampaigns(ctx context.Context, businessID int64, status string) ([]domain.Campaign, error) {
	return s.repo.ListCampaigns(ctx, businessID, status)
}

func (s *Service) GetCampaign(ctx context.Context, businessID, campaignID int64) (domain.Campaign, error) {
	return s.repo.GetCampaignByID(ctx, businessID, campaignID)
}

func (s *Service) CreateCampaign(ctx context.Context, input domain.CampaignInput) (int64, error) {
	return s.repo.CreateCampaign(ctx, input)
}

func (s *Service) UpdateCampaign(ctx context.Context, input domain.CampaignUpdateInput) error {
	return s.repo.UpdateCampaign(ctx, input)
}

func (s *Service) UpdateCampaignStatus(ctx context.Context, businessID, campaignID int64, status string) error {
	return s.repo.UpdateCampaignStatus(ctx, businessID, campaignID, status)
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

func (s *Service) GetDashboardStats(ctx context.Context, businessID int64) (domain.DashboardStats, error) {
	cacheKey := fmt.Sprintf("crm:dashboard:%d", businessID)

	return cacher.Of[domain.DashboardStats]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) (domain.DashboardStats, error) {
		return s.repo.GetDashboardStats(ctx, businessID)
	}, cacher.WithExpiry(ctx, dashboardCacheTTL))
}

func (s *Service) invalidateDashboardCache(ctx context.Context, businessID int64) {
	cacheKey := fmt.Sprintf("crm:dashboard:%d", businessID)
	cacher.Of[domain.DashboardStats]().Invalidate(ctx, []string{cacheKey})
}

// -- Timeline ----------------------------------------------------------------

func (s *Service) GetCustomerTimeline(ctx context.Context, businessID, customerID int64) ([]domain.TimelineEntry, error) {
	return s.repo.GetCustomerTimeline(ctx, businessID, customerID)
}

func (s *Service) ListCustomerTimelinePage(ctx context.Context, businessID, customerID int64, cursorCreatedAt *time.Time, cursorID int64, pageSize int) (domain.TimelinePage, error) {
	const maxPageSize = 100
	if pageSize <= 0 || pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	entries, err := s.repo.ListCustomerTimelinePage(ctx, businessID, customerID, cursorCreatedAt, cursorID, pageSize+1)
	if err != nil {
		return domain.TimelinePage{}, err
	}
	page := domain.TimelinePage{Entries: entries}
	if len(entries) > pageSize {
		page.Entries = entries[:pageSize]
		last := page.Entries[len(page.Entries)-1]
		page.NextCursor = fmt.Sprintf("%d:%d", last.CreatedAt.UTC().UnixNano(), last.ID)
	}
	return page, nil
}

// -- Customer Detail with Loyalty --------------------------------------------

func (s *Service) GetCustomerDetail(ctx context.Context, businessID, customerID int64) (domain.CustomerDetail, error) {
	customer, err := s.repo.GetCustomerByID(ctx, businessID, customerID)
	if err != nil {
		return domain.CustomerDetail{}, err
	}

	txs, _, txErr := s.repo.ListLoyaltyTransactions(ctx, businessID, customerID, 1, 10)
	if txErr != nil {
		logger.Ctx(ctx).Error("Failed to get loyalty transactions for detail", logger.Error(txErr))
		txs = []domain.LoyaltyTransaction{}
	}

	timeline, tlErr := s.repo.GetCustomerTimeline(ctx, businessID, customerID)
	if tlErr != nil {
		logger.Ctx(ctx).Error("Failed to get timeline for detail", logger.Error(tlErr))
		timeline = []domain.TimelineEntry{}
	}

	return domain.CustomerDetail{
		Customer:            customer,
		LoyaltyTransactions: txs,
		RecentVisits:        timeline,
	}, nil
}

// -- Import ------------------------------------------------------------------

func (s *Service) ImportCustomers(ctx context.Context, businessID, ownerID int64, items []domain.CustomerImportItem, dedupe bool) (domain.CustomerImportResult, error) {
	imported, failed, err := s.repo.BulkCreateCustomers(ctx, businessID, ownerID, items, dedupe)
	if err != nil {
		return domain.CustomerImportResult{}, err
	}

	invalidateCustomerScope(ctx, businessID)
	s.invalidateDashboardCache(ctx, businessID)

	return domain.CustomerImportResult{
		TotalProcessed: len(items),
		Imported:       imported,
		Failed:         failed,
	}, nil
}

// -- Export (GDPR) -----------------------------------------------------------

func (s *Service) ExportCustomerData(ctx context.Context, businessID, customerID int64) (domain.CustomerExport, error) {
	return s.repo.ExportCustomerData(ctx, businessID, customerID)
}

// -- Merge -------------------------------------------------------------------

func (s *Service) MergeCustomers(ctx context.Context, input domain.MergeCustomersInput) error {
	if err := s.repo.MergeCustomers(ctx, input); err != nil {
		return err
	}

	// Without this, the merged-away secondary keeps being served from the
	// per-customer cache even though the DB no longer has it.
	invalidateCustomerScope(ctx, input.BusinessID)
	s.invalidateDashboardCache(ctx, input.BusinessID)

	s.insertTimeline(ctx, input.BusinessID, input.OwnerID, input.PrimaryID,
		"note", fmt.Sprintf("Merged customer #%d into this profile", input.SecondaryID),
		map[string]any{"merged_customer_id": input.SecondaryID})

	return nil
}

// insertTimeline best-effort records an entry in the customer timeline. The
// timeline is auxiliary (profile activity feed + GDPR export); a failed write
// must not fail the underlying business operation.
func (s *Service) insertTimeline(ctx context.Context, businessID, ownerID, customerID int64, entryType, description string, metadata map[string]any) {
	raw, err := json.Marshal(metadata)
	if err != nil {
		raw = []byte("{}")
	}
	if err := s.repo.InsertTimelineEntry(ctx, domain.TimelineEntry{
		OwnerID:     ownerID,
		BusinessID:  businessID,
		CustomerID:  customerID,
		Type:        entryType,
		Description: description,
		Metadata:    raw,
	}); err != nil {
		logger.Ctx(ctx).Warn("failed to write customer timeline entry", logger.Error(err))
	}
}

// -- Loyalty Tiers -----------------------------------------------------------

func (s *Service) ListLoyaltyTiers(ctx context.Context, businessID int64) ([]domain.LoyaltyTier, error) {
	return s.repo.ListLoyaltyTiers(ctx, businessID)
}

func (s *Service) CreateLoyaltyTier(ctx context.Context, input domain.LoyaltyTierInput) (int64, error) {
	return s.repo.CreateLoyaltyTier(ctx, input)
}

func (s *Service) UpdateLoyaltyTier(ctx context.Context, input domain.LoyaltyTierUpdateInput) error {
	return s.repo.UpdateLoyaltyTier(ctx, input)
}

func (s *Service) DeleteLoyaltyTier(ctx context.Context, businessID, tierID int64) error {
	return s.repo.DeleteLoyaltyTier(ctx, businessID, tierID)
}

// -- Campaign Stats ----------------------------------------------------------

func (s *Service) GetCampaignStats(ctx context.Context, businessID, campaignID int64) (domain.CampaignStats, error) {
	return s.repo.GetCampaignStats(ctx, businessID, campaignID)
}

// -- Campaign Detail ---------------------------------------------------------

func (s *Service) GetCampaignDetail(ctx context.Context, businessID, campaignID int64) (domain.CampaignDetail, error) {
	campaign, err := s.repo.GetCampaignByID(ctx, businessID, campaignID)
	if err != nil {
		return domain.CampaignDetail{}, err
	}

	stats, sErr := s.repo.GetCampaignStats(ctx, businessID, campaignID)
	if sErr != nil {
		logger.Ctx(ctx).Error("Failed to get campaign stats for detail", logger.Error(sErr))
		stats = domain.CampaignStats{CampaignID: campaignID}
	}

	return domain.CampaignDetail{
		Campaign: campaign,
		Stats:    stats,
	}, nil
}

// -- Segment Detail ----------------------------------------------------------

func (s *Service) GetSegmentDetail(ctx context.Context, businessID, segmentID int64) (domain.SegmentDetail, error) {
	segment, err := s.repo.GetSegmentByID(ctx, businessID, segmentID)
	if err != nil {
		return domain.SegmentDetail{}, err
	}

	customers, _, cErr := s.repo.GetSegmentCustomers(ctx, businessID, segmentID, 1, 50)
	if cErr != nil {
		logger.Ctx(ctx).Error("Failed to get segment customers for detail", logger.Error(cErr))
		customers = []domain.Customer{}
	}

	return domain.SegmentDetail{
		Segment:   segment,
		Customers: customers,
	}, nil
}

// -- Segment Customers -------------------------------------------------------

func (s *Service) GetSegmentCustomers(ctx context.Context, businessID, segmentID int64, page, pageSize int) ([]domain.Customer, int, error) {
	return s.repo.GetSegmentCustomers(ctx, businessID, segmentID, page, pageSize)
}
