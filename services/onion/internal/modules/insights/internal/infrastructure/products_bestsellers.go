package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/platform/logger"
)

// BestSellerRow is one row in the top-N-by-units-sold ranking.
type BestSellerRow struct {
	ProductID         int64
	Name              string
	UnitsSold         float64
	RevenueMinorUnits int64
}

// GetBestSellers returns the top `limit` products by units sold in the
// last 30 days. Units is a DECIMAL in the source table (pos allows
// fractional kg/litres) so we carry float64 through.
func (r *Repository) GetBestSellers(ctx context.Context, businessID int64, limit int32) ([]BestSellerRow, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	cutoff := nowFn().AddDate(0, 0, -30)
	// cross-schema read from erp_products + pos_orders + pos_order_items — 30d best-seller ranking
	const q = `
		SELECT p.id,
		       p.name,
		       COALESCE(SUM(oi.quantity), 0) AS units,
		       COALESCE(SUM(oi.total), 0)    AS revenue
		FROM erp_products p
		JOIN pos_order_items oi ON oi.menu_item_id = p.id
		JOIN pos_orders o       ON o.id = oi.order_id AND o.business_id = p.business_id
		WHERE p.business_id = ?
		  AND o.created_at >= ?
		GROUP BY p.id, p.name
		ORDER BY units DESC
		LIMIT ?
	`
	rows, err := r.pool.QueryContext(ctx, q, businessID, cutoff, limit)
	if err != nil {
		logger.Ctx(ctx).Error("insights: best sellers query failed", logger.Error(err))
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Warn("insights: rows.Close failed", logger.Error(closeErr))
		}
	}()
	out := make([]BestSellerRow, 0, limit)
	for rows.Next() {
		var row BestSellerRow
		var units, revenue interface{}
		if scanErr := rows.Scan(&row.ProductID, &row.Name, &units, &revenue); scanErr != nil {
			logger.Ctx(ctx).Error("insights: best sellers scan failed", logger.Error(scanErr))
			return nil, scanErr
		}
		row.UnitsSold = toFloat(units)
		row.RevenueMinorUnits = decimalToMinorUnits(revenue)
		out = append(out, row)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	return out, nil
}
