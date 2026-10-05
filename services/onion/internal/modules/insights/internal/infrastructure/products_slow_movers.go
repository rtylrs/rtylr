package infrastructure

import (
	"context"
	"database/sql"
	"github.com/voxire/onion/internal/modules/insights/internal/domain"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type SlowMoverRow = domain.SlowMoverRow

func (r *Repository) GetSlowMovers(ctx context.Context, businessID int64, days int32) ([]SlowMoverRow, error) {
	if days <= 0 {
		days = 60
	}
	if days > 365 {
		days = 365
	}
	cutoff := nowFn().AddDate(0, 0, -int(days))
	// cross-schema read from erp_products + pos_order_items + pos_orders — slow-mover detection
	rows, err := r.queries.GetSlowMovers(ctx, sqlc.GetSlowMoversParams{
		BusinessID:   businessID,
		CreatedAt:    sql.NullTime{Time: cutoff, Valid: true},
		BusinessID_2: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: slow movers query failed", logger.Error(err))
		return nil, err
	}
	out := make([]SlowMoverRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, SlowMoverRow{
			ProductID: row.ID, Name: row.Name, StockQuantity: toFloat(row.StockQuantity),
		})
	}
	return out, nil
}
