package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/customers/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	dashboardCacheTTL = 5 * time.Minute
	loyaltyCacheTTL   = 30 * time.Minute
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

// ─── Customers ────────────────────────────────────────────────────────────────

func (s *Service) ListCustomers(ctx context.Context, filter domain.CustomerListFilter) (domain.CustomerListResult, error) {
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
}

func (s *Service) GetCustomer(ctx context.Context, ownerID, customerID int64) (domain.Customer, error) {
	return s.repo.GetCustomerByID(ctx, ownerID, customerID)
}

func (s *Service) CreateCustomer(ctx context.Context, input domain.CustomerInput) (int64, error) {
	return s.repo.CreateCustomer(ctx, input)
}

func (s *Service) UpdateCustomer(ctx context.Context, input domain.CustomerUpdateInput) error {
	return s.repo.UpdateCustomer(ctx, input)
}

func (s *Service) DeleteCustomer(ctx context.Context, ownerID, customerID int64) error {
	return s.repo.DeleteCustomer(ctx, ownerID, customerID)
}

func (s *Service) RecordVisit(ctx context.Context, input domain.RecordVisitInput) error {
	if err := s.repo.RecordVisit(ctx, input.OwnerID, input.CustomerID, input.Amount); err != nil {
		return err
	}

	if input.AwardPoints && input.Amount > 0 {
		cfg, err := s.GetLoyaltyConfig(ctx, input.OwnerID)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to get loyalty config for visit award", logger.Error(err))

			return nil
		}

		if cfg.IsEnabled && cfg.PointsPerCurrencyUnit > 0 {
			points := int(input.Amount * cfg.PointsPerCurrencyUnit)
			if points > 0 {
				if awardErr := s.AwardPoints(ctx, domain.AwardPointsInput{
					OwnerID:     input.OwnerID,
					CustomerID:  input.CustomerID,
					Points:      points,
					Description: fmt.Sprintf("Points earned for visit (spend: %.2f)", input.Amount),
				}); awardErr != nil {
					logger.Ctx(ctx).Error("Failed to auto-award points on visit", logger.Error(awardErr))
				}
			}
		}
	}

	s.invalidateDashboardCache(ctx, input.OwnerID)

	return nil
}

// ─── Loyalty ──────────────────────────────────────────────────────────────────

func (s *Service) AwardPoints(ctx context.Context, input domain.AwardPointsInput) error {
	customer, err := s.repo.GetCustomerByID(ctx, input.OwnerID, input.CustomerID)
	if err != nil {
		return err
	}

	balanceBefore := customer.LoyaltyPoints
	balanceAfter := balanceBefore + input.Points

	tx := domain.LoyaltyTransaction{
		OwnerID:       input.OwnerID,
		CustomerID:    input.CustomerID,
		Type:          "earn",
		Points:        input.Points,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		Description:   input.Description,
		ReferenceID:   input.ReferenceID,
		ReferenceType: input.ReferenceType,
	}

	if err := s.repo.CreateLoyaltyTransaction(ctx, tx); err != nil {
		return err
	}

	cfg, cfgErr := s.GetLoyaltyConfig(ctx, input.OwnerID)
	if cfgErr != nil {
		logger.Ctx(ctx).Error("Failed to get loyalty config, using default tier", logger.Error(cfgErr))
	}

	newTier := RecalculateTier(balanceAfter, cfg)

	return s.repo.UpdateCustomerLoyalty(ctx, input.CustomerID, balanceAfter, newTier)
}

func (s *Service) RedeemPoints(ctx context.Context, input domain.RedeemPointsInput) error {
	customer, err := s.repo.GetCustomerByID(ctx, input.OwnerID, input.CustomerID)
	if err != nil {
		return err
	}

	if customer.LoyaltyPoints < input.Points {
		return domain.ErrInsufficientPoints
	}

	balanceBefore := customer.LoyaltyPoints
	balanceAfter := balanceBefore - input.Points

	tx := domain.LoyaltyTransaction{
		OwnerID:       input.OwnerID,
		CustomerID:    input.CustomerID,
		Type:          "redeem",
		Points:        input.Points,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		Description:   input.Description,
		ReferenceID:   input.ReferenceID,
		ReferenceType: input.ReferenceType,
	}

	if err := s.repo.CreateLoyaltyTransaction(ctx, tx); err != nil {
		return err
	}

	cfg, cfgErr := s.GetLoyaltyConfig(ctx, input.OwnerID)
	if cfgErr != nil {
		logger.Ctx(ctx).Error("Failed to get loyalty config, using default tier", logger.Error(cfgErr))
	}

	newTier := RecalculateTier(balanceAfter, cfg)

	return s.repo.UpdateCustomerLoyalty(ctx, input.CustomerID, balanceAfter, newTier)
}

func (s *Service) ListLoyaltyTransactions(ctx context.Context, ownerID, customerID int64, page, pageSize int) ([]domain.LoyaltyTransaction, int, error) {
	return s.repo.ListLoyaltyTransactions(ctx, ownerID, customerID, page, pageSize)
}

func (s *Service) GetLoyaltyConfig(ctx context.Context, ownerID int64) (domain.LoyaltyConfig, error) {
	cacheKey := fmt.Sprintf("crm:loyalty_config:%d", ownerID)

	return cacher.Of[domain.LoyaltyConfig]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) (domain.LoyaltyConfig, error) {
		return s.repo.GetLoyaltyConfig(ctx, ownerID)
	}, cacher.WithExpiry(ctx, loyaltyCacheTTL))
}

func (s *Service) UpsertLoyaltyConfig(ctx context.Context, input domain.LoyaltyConfigInput) error {
	if err := s.repo.UpsertLoyaltyConfig(ctx, input); err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("crm:loyalty_config:%d", input.OwnerID)
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

func (s *Service) ListSegments(ctx context.Context, ownerID int64) ([]domain.Segment, error) {
	return s.repo.ListSegments(ctx, ownerID)
}

func (s *Service) GetSegment(ctx context.Context, ownerID, segmentID int64) (domain.Segment, error) {
	return s.repo.GetSegmentByID(ctx, ownerID, segmentID)
}

func (s *Service) CreateSegment(ctx context.Context, input domain.SegmentInput) (int64, error) {
	return s.repo.CreateSegment(ctx, input)
}

func (s *Service) UpdateSegment(ctx context.Context, input domain.SegmentUpdateInput) error {
	return s.repo.UpdateSegment(ctx, input)
}

func (s *Service) DeleteSegment(ctx context.Context, ownerID, segmentID int64) error {
	return s.repo.DeleteSegment(ctx, ownerID, segmentID)
}

// ─── Campaigns ────────────────────────────────────────────────────────────────

func (s *Service) ListCampaigns(ctx context.Context, ownerID int64, status string) ([]domain.Campaign, error) {
	return s.repo.ListCampaigns(ctx, ownerID, status)
}

func (s *Service) GetCampaign(ctx context.Context, ownerID, campaignID int64) (domain.Campaign, error) {
	return s.repo.GetCampaignByID(ctx, ownerID, campaignID)
}

func (s *Service) CreateCampaign(ctx context.Context, input domain.CampaignInput) (int64, error) {
	return s.repo.CreateCampaign(ctx, input)
}

func (s *Service) UpdateCampaign(ctx context.Context, input domain.CampaignUpdateInput) error {
	return s.repo.UpdateCampaign(ctx, input)
}

func (s *Service) UpdateCampaignStatus(ctx context.Context, ownerID, campaignID int64, status string) error {
	return s.repo.UpdateCampaignStatus(ctx, ownerID, campaignID, status)
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

func (s *Service) GetDashboardStats(ctx context.Context, ownerID int64) (domain.DashboardStats, error) {
	cacheKey := fmt.Sprintf("crm:dashboard:%d", ownerID)

	return cacher.Of[domain.DashboardStats]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) (domain.DashboardStats, error) {
		return s.repo.GetDashboardStats(ctx, ownerID)
	}, cacher.WithExpiry(ctx, dashboardCacheTTL))
}

func (s *Service) invalidateDashboardCache(ctx context.Context, ownerID int64) {
	cacheKey := fmt.Sprintf("crm:dashboard:%d", ownerID)
	cacher.Of[domain.DashboardStats]().Invalidate(ctx, []string{cacheKey})
}

// -- Timeline ----------------------------------------------------------------

func (s *Service) GetCustomerTimeline(ctx context.Context, ownerID, customerID int64) ([]domain.TimelineEntry, error) {
	return s.repo.GetCustomerTimeline(ctx, ownerID, customerID)
}

// -- Customer Detail with Loyalty --------------------------------------------

func (s *Service) GetCustomerDetail(ctx context.Context, ownerID, customerID int64) (domain.CustomerDetail, error) {
	customer, err := s.repo.GetCustomerByID(ctx, ownerID, customerID)
	if err != nil {
		return domain.CustomerDetail{}, err
	}

	txs, _, txErr := s.repo.ListLoyaltyTransactions(ctx, ownerID, customerID, 1, 10)
	if txErr != nil {
		logger.Ctx(ctx).Error("Failed to get loyalty transactions for detail", logger.Error(txErr))
		txs = []domain.LoyaltyTransaction{}
	}

	timeline, tlErr := s.repo.GetCustomerTimeline(ctx, ownerID, customerID)
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

func (s *Service) ImportCustomers(ctx context.Context, ownerID int64, items []domain.CustomerImportItem) (domain.CustomerImportResult, error) {
	imported, failed, err := s.repo.BulkCreateCustomers(ctx, ownerID, items)
	if err != nil {
		return domain.CustomerImportResult{}, err
	}

	s.invalidateDashboardCache(ctx, ownerID)

	return domain.CustomerImportResult{
		TotalProcessed: len(items),
		Imported:       imported,
		Failed:         failed,
	}, nil
}

// -- Export (GDPR) -----------------------------------------------------------

func (s *Service) ExportCustomerData(ctx context.Context, ownerID, customerID int64) (domain.CustomerExport, error) {
	return s.repo.ExportCustomerData(ctx, ownerID, customerID)
}

// -- Merge -------------------------------------------------------------------

func (s *Service) MergeCustomers(ctx context.Context, input domain.MergeCustomersInput) error {
	if err := s.repo.MergeCustomers(ctx, input); err != nil {
		return err
	}

	s.invalidateDashboardCache(ctx, input.OwnerID)

	return nil
}

// -- Loyalty Tiers -----------------------------------------------------------

func (s *Service) ListLoyaltyTiers(ctx context.Context, ownerID int64) ([]domain.LoyaltyTier, error) {
	return s.repo.ListLoyaltyTiers(ctx, ownerID)
}

func (s *Service) CreateLoyaltyTier(ctx context.Context, input domain.LoyaltyTierInput) (int64, error) {
	return s.repo.CreateLoyaltyTier(ctx, input)
}

func (s *Service) UpdateLoyaltyTier(ctx context.Context, input domain.LoyaltyTierUpdateInput) error {
	return s.repo.UpdateLoyaltyTier(ctx, input)
}

func (s *Service) DeleteLoyaltyTier(ctx context.Context, ownerID, tierID int64) error {
	return s.repo.DeleteLoyaltyTier(ctx, ownerID, tierID)
}

// -- Campaign Stats ----------------------------------------------------------

func (s *Service) GetCampaignStats(ctx context.Context, ownerID, campaignID int64) (domain.CampaignStats, error) {
	return s.repo.GetCampaignStats(ctx, ownerID, campaignID)
}

// -- Campaign Detail ---------------------------------------------------------

func (s *Service) GetCampaignDetail(ctx context.Context, ownerID, campaignID int64) (domain.CampaignDetail, error) {
	campaign, err := s.repo.GetCampaignByID(ctx, ownerID, campaignID)
	if err != nil {
		return domain.CampaignDetail{}, err
	}

	stats, sErr := s.repo.GetCampaignStats(ctx, ownerID, campaignID)
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

func (s *Service) GetSegmentDetail(ctx context.Context, ownerID, segmentID int64) (domain.SegmentDetail, error) {
	segment, err := s.repo.GetSegmentByID(ctx, ownerID, segmentID)
	if err != nil {
		return domain.SegmentDetail{}, err
	}

	customers, _, cErr := s.repo.GetSegmentCustomers(ctx, ownerID, segmentID, 1, 50)
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

func (s *Service) GetSegmentCustomers(ctx context.Context, ownerID, segmentID int64, page, pageSize int) ([]domain.Customer, int, error) {
	return s.repo.GetSegmentCustomers(ctx, ownerID, segmentID, page, pageSize)
}
