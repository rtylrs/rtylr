package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/logger"
)

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

	s.insertTimeline(ctx, input.BusinessID, input.OwnerID, input.PrimaryID,
		"note", fmt.Sprintf("Merged customer #%d into this profile", input.SecondaryID),
		map[string]any{"merged_customer_id": input.SecondaryID})

	return nil
}
