package application

import (
	"context"
	"encoding/json"
	"math"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

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

func normalizeLoyaltyTierPage(page, pageSize int) (int, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = domain.DefaultLoyaltyTierPageSize
	}
	if pageSize > domain.MaxLoyaltyTierPageSize {
		pageSize = domain.MaxLoyaltyTierPageSize
	}
	if page-1 > math.MaxInt32/pageSize {
		return 0, 0, domain.ErrLoyaltyTierPageOutOfRange
	}
	return page, pageSize, nil
}

func (s *Service) ListLoyaltyTiersPage(ctx context.Context, businessID int64, page, pageSize int) (domain.LoyaltyTierPage, error) {
	page, pageSize, err := normalizeLoyaltyTierPage(page, pageSize)
	if err != nil {
		return domain.LoyaltyTierPage{}, err
	}

	tiers, total, err := s.repo.ListLoyaltyTiersPage(ctx, businessID, page, pageSize)
	if err != nil {
		return domain.LoyaltyTierPage{}, err
	}
	end := int64(page-1)*int64(pageSize) + int64(len(tiers))
	return domain.LoyaltyTierPage{
		Tiers:    tiers,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasMore:  end < total,
	}, nil
}

// ListLoyaltyTiers retains the unpaginated service API for existing internal
// callers. The repository implementation follows every page, so this method
// remains complete even when a business has more than one database page.
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
