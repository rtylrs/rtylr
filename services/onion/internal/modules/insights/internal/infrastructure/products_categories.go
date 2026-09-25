package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type CategorySalesRow struct {
	CategoryID        int64
	CategoryName      string
	UnitsSold         float64
	RevenueMinorUnits int64
}

func (r *Repository) GetSalesByCategory(ctx context.Context, businessID int64) ([]CategorySalesRow, error) {
	thisFrom, thisTo, _, _ := mtdRanges(nowFn())
	// cross-schema read from erp_categories + erp_products + pos_order_items + pos_orders — MTD category sales
	rows, err := r.queries.GetSalesByCategory(ctx, sqlc.GetSalesByCategoryParams{
		BusinessID:  businessID,
		CreatedAt:   sql.NullTime{Time: thisFrom, Valid: true},
		CreatedAt_2: sql.NullTime{Time: thisTo, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: category sales query failed", logger.Error(err))
		return nil, err
	}
	out := make([]CategorySalesRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, CategorySalesRow{
			CategoryID: row.CategoryID, CategoryName: row.CategoryName,
			UnitsSold: toFloat(row.Units), RevenueMinorUnits: decimalToMinorUnits(row.Revenue),
		})
	}
	return out, nil
}
