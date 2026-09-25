package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/customers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/segmentrules"
)

func normalizeLoyaltyTierRepositoryPage(page, pageSize int) (int, int, error) {
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

func (r *Repository) ListLoyaltyTiersPage(ctx context.Context, businessID int64, page, pageSize int) ([]domain.LoyaltyTier, int64, error) {
	page, pageSize, err := normalizeLoyaltyTierRepositoryPage(page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	total, err := r.queries.CountLoyaltyTiers(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count loyalty tiers", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListLoyaltyTiers(ctx, sqlc.ListLoyaltyTiersParams{
		BusinessID: businessID,
		Limit:      int32(pageSize),
		Offset:     int32((page - 1) * pageSize),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list loyalty tiers", logger.Error(err))
		return nil, 0, err
	}

	tiers := make([]domain.LoyaltyTier, 0, len(rows))
	for _, row := range rows {
		dp := parseDecimal(row.DiscountPct.String)
		t := domain.LoyaltyTier{
			ID:          row.ID,
			OwnerID:     row.OwnerID,
			BusinessID:  row.BusinessID,
			Name:        row.Name,
			MinPoints:   int(row.MinPoints.Int32),
			DiscountPct: dp,
			SortOrder:   int(row.SortOrder.Int32),
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		}
		if row.Perks != nil {
			t.Perks = []byte(row.Perks)
		} else {
			t.Perks = []byte("[]")
		}
		tiers = append(tiers, t)
	}

	return tiers, total, nil
}

// ListLoyaltyTiers keeps the original complete-list API for callers that do
// not have pagination metadata. It walks the bounded query instead of silently
// treating the first 100 rows as the complete business catalogue.
func (r *Repository) ListLoyaltyTiers(ctx context.Context, businessID int64) ([]domain.LoyaltyTier, error) {
	const pageSize = 100
	all := make([]domain.LoyaltyTier, 0)

	for page := 1; ; page++ {
		rows, total, err := r.ListLoyaltyTiersPage(ctx, businessID, page, pageSize)
		if err != nil {
			return nil, err
		}
		all = append(all, rows...)
		if int64(len(all)) >= total || len(rows) < pageSize {
			return all, nil
		}
	}
}

func (r *Repository) CreateLoyaltyTier(ctx context.Context, input domain.LoyaltyTierInput) (int64, error) {
	result, err := r.queries.CreateLoyaltyTier(ctx, sqlc.CreateLoyaltyTierParams{
		OwnerID:     input.OwnerID,
		BusinessID:  input.BusinessID,
		Name:        input.Name,
		MinPoints:   sql.NullInt32{Int32: int32(input.MinPoints), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		DiscountPct: sql.NullString{String: strconv.FormatFloat(input.DiscountPct, 'f', 2, 64), Valid: true},
		Perks:       json.RawMessage(input.Perks),
		SortOrder:   sql.NullInt32{Int32: int32(input.SortOrder), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create loyalty tier", logger.Error(err))

		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) UpdateLoyaltyTier(ctx context.Context, input domain.LoyaltyTierUpdateInput) error {
	if input.BusinessID == 0 {
		input.BusinessID = input.OwnerID
	}

	result, err := r.queries.UpdateLoyaltyTier(ctx, sqlc.UpdateLoyaltyTierParams{
		Name:        input.Name,
		MinPoints:   sql.NullInt32{Int32: int32(input.MinPoints), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		DiscountPct: sql.NullString{String: strconv.FormatFloat(input.DiscountPct, 'f', 2, 64), Valid: true},
		Perks:       json.RawMessage(input.Perks),
		SortOrder:   sql.NullInt32{Int32: int32(input.SortOrder), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		ID:          input.ID,
		BusinessID:  input.BusinessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update loyalty tier", logger.Error(err))

		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrTierNotFound
	}

	return nil
}

func (r *Repository) DeleteLoyaltyTier(ctx context.Context, businessID, tierID int64) error {
	tier, err := r.queries.GetLoyaltyTierByID(ctx, sqlc.GetLoyaltyTierByIDParams{
		ID:         tierID,
		BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrTierNotFound
		}

		logger.Ctx(ctx).Error("Failed to look up loyalty tier", logger.Error(err), logger.Any("business_id", businessID), logger.Any("tier_id", tierID))

		return err
	}

	// cross-schema read from crm_customers — active members on this tier
	count, err := r.queries.CountActiveMembersOnTier(ctx, sqlc.CountActiveMembersOnTierParams{
		BusinessID:  businessID,
		LoyaltyTier: sqlc.NullCrmCustomersLoyaltyTier{CrmCustomersLoyaltyTier: sqlc.CrmCustomersLoyaltyTier(strings.ToLower(strings.TrimSpace(tier.Name))), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count active members on tier", logger.Error(err), logger.Any("business_id", businessID), logger.Any("tier_id", tierID))

		return err
	}

	if count > 0 {
		return domain.ErrTierInUse
	}

	result, err := r.queries.DeleteLoyaltyTier(ctx, sqlc.DeleteLoyaltyTierParams{
		ID:         tierID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete loyalty tier", logger.Error(err), logger.Any("business_id", businessID), logger.Any("tier_id", tierID))

		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrTierNotFound
	}

	return nil
}

// ─── Campaign Stats ─────────────────────────────────────────────────────────

func (r *Repository) GetCampaignStats(ctx context.Context, businessID, campaignID int64) (domain.CampaignStats, error) {
	campaign, err := r.GetCampaignByID(ctx, businessID, campaignID)
	if err != nil {
		return domain.CampaignStats{}, err
	}

	var stats domain.CampaignStats
	stats.CampaignID = campaign.ID
	stats.AudienceCount = campaign.AudienceCount
	stats.SentCount = campaign.SentCount
	stats.DeliveredCount = campaign.DeliveredCount
	stats.OpenedCount = campaign.OpenedCount
	stats.ClickedCount = campaign.ClickedCount

	sendStats, err := r.queries.GetCampaignSendStats(ctx, sqlc.GetCampaignSendStatsParams{
		CampaignID: campaignID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get campaign send stats", logger.Error(err))
	} else {
		if v, ok := toInt(sendStats.BouncedCount); ok {
			stats.BouncedCount = v
		}
		if v, ok := toInt(sendStats.FailedCount); ok {
			stats.FailedCount = v
		}
	}

	return stats, nil
}

// ─── Segment Customers ─────────────────────────────────────────────────────

func (r *Repository) GetSegmentCustomers(ctx context.Context, businessID, segmentID int64, page, pageSize int) ([]domain.Customer, int, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	segment, err := r.GetSegmentByID(ctx, businessID, segmentID)
	if err != nil {
		return nil, 0, err
	}

	conds := sq.And{
		sq.Eq{"business_id": businessID},
		sq.Eq{"is_active": true},
	}
	// Real rule-based membership: without this, every segment returned all
	// active customers. Segments with no (valid) rules keep the legacy
	// "all active customers" behavior.
	if frag, args := segmentrules.Filter(segment.Rules); frag != "" {
		conds = append(conds, sq.Expr(frag, args...))
	}

	countSQL, countArgs, err := sq.Select("COUNT(*)").
		From("crm_customers").
		Where(conds).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.pool.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		logger.Ctx(ctx).Error("Failed to count segment customers", logger.Error(err))

		return nil, 0, err
	}

	listSQL, listArgs, err := sq.Select(
		"id", "owner_id", "business_id", "first_name", "last_name", "email", "phone", "date_of_birth", "gender",
		"address", "city", "country", "loyalty_points", "total_spent", "visit_count", "last_visit_at",
		"loyalty_tier", "tags", "notes", "consent", "is_active", "created_at", "updated_at",
	).
		From("crm_customers").
		Where(conds).
		OrderBy("created_at DESC").
		Limit(uint64(pageSize)).
		Offset(uint64(offset)). // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.QueryContext(ctx, listSQL, listArgs...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get segment customers", logger.Error(err))

		return nil, 0, err
	}
	defer rows.Close()

	customers := make([]domain.Customer, 0, pageSize)

	for rows.Next() {
		c, scanErr := scanCustomerRaw(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		customers = append(customers, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────
