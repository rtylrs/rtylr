package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/voxire/onion/internal/modules/campaigns"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/customers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/modules/segments"
	"github.com/voxire/onion/internal/platform/logger"
)

func (r *Repository) GetLoyaltyConfig(ctx context.Context, businessID int64) (domain.LoyaltyConfig, error) {
	row, err := r.queries.GetLoyaltyConfig(ctx, businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.LoyaltyConfig{}, nil
		}

		logger.Ctx(ctx).Error("Failed to get loyalty config", logger.Error(err))

		return domain.LoyaltyConfig{}, err
	}

	ppcu := parseDecimal(row.PointsPerCurrencyUnit.String)
	cpp := parseDecimal(row.CurrencyPerPoint.String)
	bdp := parseDecimal(row.BronzeDiscountPct.String)
	sdp := parseDecimal(row.SilverDiscountPct.String)
	gdp := parseDecimal(row.GoldDiscountPct.String)
	pdp := parseDecimal(row.PlatinumDiscountPct.String)

	return domain.LoyaltyConfig{
		ID:                    row.ID,
		OwnerID:               row.OwnerID,
		BusinessID:            row.BusinessID,
		IsEnabled:             row.IsEnabled.Bool,
		PointsPerCurrencyUnit: ppcu,
		CurrencyPerPoint:      cpp,
		BronzeMin:             int(row.BronzeMin.Int32),
		SilverMin:             int(row.SilverMin.Int32),
		GoldMin:               int(row.GoldMin.Int32),
		PlatinumMin:           int(row.PlatinumMin.Int32),
		BronzeDiscountPct:     bdp,
		SilverDiscountPct:     sdp,
		GoldDiscountPct:       gdp,
		PlatinumDiscountPct:   pdp,
		PointsExpiryDays:      int(row.PointsExpiryDays.Int32),
		UpdatedAt:             row.UpdatedAt.Time,
	}, nil
}

func (r *Repository) UpsertLoyaltyConfig(ctx context.Context, input domain.LoyaltyConfigInput) error {
	err := r.queries.UpsertLoyaltyConfig(ctx, sqlc.UpsertLoyaltyConfigParams{
		OwnerID:               input.OwnerID,
		BusinessID:            input.BusinessID,
		IsEnabled:             sql.NullBool{Bool: input.IsEnabled, Valid: true},
		PointsPerCurrencyUnit: sql.NullString{String: strconv.FormatFloat(input.PointsPerCurrencyUnit, 'f', 2, 64), Valid: true},
		CurrencyPerPoint:      sql.NullString{String: strconv.FormatFloat(input.CurrencyPerPoint, 'f', 4, 64), Valid: true},
		BronzeMin:             sql.NullInt32{Int32: int32(input.BronzeMin), Valid: true},   // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		SilverMin:             sql.NullInt32{Int32: int32(input.SilverMin), Valid: true},   // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		GoldMin:               sql.NullInt32{Int32: int32(input.GoldMin), Valid: true},     // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PlatinumMin:           sql.NullInt32{Int32: int32(input.PlatinumMin), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		BronzeDiscountPct:     sql.NullString{String: strconv.FormatFloat(input.BronzeDiscountPct, 'f', 2, 64), Valid: true},
		SilverDiscountPct:     sql.NullString{String: strconv.FormatFloat(input.SilverDiscountPct, 'f', 2, 64), Valid: true},
		GoldDiscountPct:       sql.NullString{String: strconv.FormatFloat(input.GoldDiscountPct, 'f', 2, 64), Valid: true},
		PlatinumDiscountPct:   sql.NullString{String: strconv.FormatFloat(input.PlatinumDiscountPct, 'f', 2, 64), Valid: true},
		PointsExpiryDays:      sql.NullInt32{Int32: int32(input.PointsExpiryDays), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upsert loyalty config", logger.Error(err))
	}

	return err
}

// ─── Segments ─────────────────────────────────────────────────────────────────
//
// Segment persistence belongs to the segments module. These compatibility
// methods keep the existing customer routes stable while delegating to its
// public facade rather than maintaining another SQLC query surface.

func (r *Repository) ListSegments(ctx context.Context, businessID int64) ([]domain.Segment, error) {
	rows, err := r.segments.List(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list segments", logger.Error(err))

		return nil, err
	}

	segments := make([]domain.Segment, 0, len(rows))

	for _, row := range rows {
		segments = append(segments, segmentFromCanonical(row))
	}

	return segments, nil
}

func (r *Repository) GetSegmentByID(ctx context.Context, businessID, segmentID int64) (domain.Segment, error) {
	row, err := r.segments.Get(ctx, businessID, segmentID)
	if err != nil {
		if errors.Is(err, segments.ErrNotFound) {
			return domain.Segment{}, domain.ErrSegmentNotFound
		}

		return domain.Segment{}, err
	}

	return segmentFromCanonical(row), nil
}

func (r *Repository) CreateSegment(ctx context.Context, input domain.SegmentInput) (int64, error) {
	id, err := r.segments.Create(ctx, segments.SegmentInput{
		OwnerID: input.OwnerID, BusinessID: input.BusinessID, Name: input.Name, Description: input.Description, Rules: input.Rules,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create segment", logger.Error(err))

		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdateSegment(ctx context.Context, input domain.SegmentUpdateInput) error {
	if input.BusinessID == 0 {
		input.BusinessID = input.OwnerID
	}

	err := r.segments.Update(ctx, segments.SegmentUpdateInput{
		ID: input.ID, OwnerID: input.OwnerID, BusinessID: input.BusinessID, Name: input.Name, Description: input.Description, Rules: input.Rules,
	})
	if err != nil {
		if errors.Is(err, segments.ErrNotFound) {
			return domain.ErrSegmentNotFound
		}
		logger.Ctx(ctx).Error("Failed to update segment", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) DeleteSegment(ctx context.Context, businessID, segmentID int64) error {
	err := r.segments.Delete(ctx, businessID, segmentID)
	if err != nil {
		if errors.Is(err, segments.ErrNotFound) {
			return domain.ErrSegmentNotFound
		}
		logger.Ctx(ctx).Error("Failed to delete segment", logger.Error(err))

		return err
	}

	return nil
}

// ─── Campaigns ────────────────────────────────────────────────────────────────
//
// Campaign persistence belongs to the campaigns module. These compatibility
// methods retain the customer API's DTOs while delegating to the canonical
// campaign facade.

func (r *Repository) ListCampaigns(ctx context.Context, businessID int64, status string) ([]domain.Campaign, error) {
	rows, err := r.campaigns.List(ctx, businessID, status)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list campaigns", logger.Error(err))

		return nil, err
	}

	campaigns := make([]domain.Campaign, 0, len(rows))

	for _, row := range rows {
		campaigns = append(campaigns, campaignFromCanonical(row))
	}

	return campaigns, nil
}

func (r *Repository) GetCampaignByID(ctx context.Context, businessID, campaignID int64) (domain.Campaign, error) {
	row, err := r.campaigns.Get(ctx, businessID, campaignID)
	if err != nil {
		if errors.Is(err, campaigns.ErrNotFound) {
			return domain.Campaign{}, domain.ErrCampaignNotFound
		}

		return domain.Campaign{}, err
	}

	return campaignFromCanonical(row), nil
}

func (r *Repository) CreateCampaign(ctx context.Context, input domain.CampaignInput) (int64, error) {
	id, err := r.campaigns.Create(ctx, campaigns.CampaignInput{
		OwnerID: input.OwnerID, BusinessID: input.BusinessID, Name: input.Name, Description: input.Description, Channel: input.Channel,
		SegmentID: input.SegmentID, Subject: input.Subject, Content: input.Content, ScheduledAt: input.ScheduledAt,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create campaign", logger.Error(err))

		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdateCampaign(ctx context.Context, input domain.CampaignUpdateInput) error {
	if input.BusinessID == 0 {
		input.BusinessID = input.OwnerID
	}

	err := r.campaigns.Update(ctx, campaigns.CampaignUpdateInput{
		ID: input.ID, OwnerID: input.OwnerID, BusinessID: input.BusinessID, Name: input.Name, Description: input.Description,
		Channel: input.Channel, SegmentID: input.SegmentID, Subject: input.Subject, Content: input.Content, ScheduledAt: input.ScheduledAt,
	})
	if err != nil {
		if errors.Is(err, campaigns.ErrNotFound) {
			return domain.ErrCampaignNotFound
		}
		logger.Ctx(ctx).Error("Failed to update campaign", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) UpdateCampaignStatus(ctx context.Context, businessID, campaignID int64, status string) error {
	err := r.campaigns.UpdateStatus(ctx, businessID, campaignID, status)
	if err != nil {
		if errors.Is(err, campaigns.ErrNotFound) {
			return domain.ErrCampaignNotFound
		}
		logger.Ctx(ctx).Error("Failed to update campaign status", logger.Error(err))

		return err
	}

	return nil
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

func (r *Repository) GetDashboardStats(ctx context.Context, businessID int64) (domain.DashboardStats, error) {
	var (
		totalCustomers int64
		newThisMonth   int64
		activeMembers  int64
		topRows        []sqlc.ListTopCustomersRow
	)

	timezone, err := r.queries.GetBusinessTimezone(ctx, businessID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("Failed to get business timezone", logger.Error(err))
		return domain.DashboardStats{}, err
	}
	location := time.UTC
	if timezone != "" {
		loaded, loadErr := time.LoadLocation(timezone)
		if loadErr != nil {
			logger.Ctx(ctx).Warn("Invalid business timezone; falling back to UTC", logger.Error(loadErr))
		} else {
			location = loaded
		}
	}

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		var err error
		totalCustomers, err = r.queries.CountActiveCustomers(groupCtx, businessID)
		if err != nil {
			logger.Ctx(groupCtx).Error("Failed to get total customers", logger.Error(err))
		}
		return err
	})
	group.Go(func() error {
		var err error
		monthStart, monthEnd := customerMonthBounds(time.Now(), location)
		newThisMonth, err = r.queries.CountNewCustomersThisMonth(groupCtx, sqlc.CountNewCustomersThisMonthParams{
			BusinessID:     businessID,
			CreatedAtStart: sql.NullTime{Time: monthStart, Valid: true},
			CreatedAtEnd:   sql.NullTime{Time: monthEnd, Valid: true},
		})
		if err != nil {
			logger.Ctx(groupCtx).Error("Failed to get new customers this month", logger.Error(err))
		}
		return err
	})
	group.Go(func() error {
		var err error
		activeMembers, err = r.queries.CountActiveLoyaltyMembers(groupCtx, businessID)
		if err != nil {
			logger.Ctx(groupCtx).Error("Failed to get active loyalty members", logger.Error(err))
		}
		return err
	})
	group.Go(func() error {
		var err error
		topRows, err = r.queries.ListTopCustomers(groupCtx, businessID)
		if err != nil {
			logger.Ctx(groupCtx).Error("Failed to get top customers", logger.Error(err))
		}
		return err
	})
	if err := group.Wait(); err != nil {
		return domain.DashboardStats{}, err
	}

	stats := domain.DashboardStats{
		TotalCustomers:       int(totalCustomers),
		NewThisMonth:         int(newThisMonth),
		ActiveLoyaltyMembers: int(activeMembers),
		TopCustomers:         make([]domain.TopCustomer, 0, len(topRows)),
	}
	for _, row := range topRows {
		spent := parseDecimal(row.TotalSpent.String)
		stats.TopCustomers = append(stats.TopCustomers, domain.TopCustomer{
			ID:            row.ID,
			FirstName:     row.FirstName,
			LastName:      row.LastName.String,
			Email:         row.Email.String,
			TotalSpent:    spent,
			LoyaltyPoints: int(row.LoyaltyPoints.Int32),
			LoyaltyTier:   string(row.LoyaltyTier.CrmCustomersLoyaltyTier),
		})
	}

	return stats, nil
}
