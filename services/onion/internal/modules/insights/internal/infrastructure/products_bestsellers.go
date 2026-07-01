package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type BestSellerRow struct {
	ProductID         int64
	Name              string
	UnitsSold         float64
	RevenueMinorUnits int64
}

func (r *Repository) GetBestSellers(ctx context.Context, businessID int64, limit int32) ([]BestSellerRow, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	cutoff := nowFn().AddDate(0, 0, -30)
	// cross-schema read from erp_products + pos_orders + pos_order_items — 30d best-seller ranking
	rows, err := r.queries.GetBestSellers(ctx, sqlc.GetBestSellersParams{
		BusinessID: businessID,
		CreatedAt:  sql.NullTime{Time: cutoff, Valid: true},
		Limit:      limit,
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: best sellers query failed", logger.Error(err))
		return nil, err
	}
	out := make([]BestSellerRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, BestSellerRow{
			ProductID: row.ID, Name: row.Name,
			UnitsSold: toFloat(row.Units), RevenueMinorUnits: decimalToMinorUnits(row.Revenue),
		})
	}
	return out, nil
}
