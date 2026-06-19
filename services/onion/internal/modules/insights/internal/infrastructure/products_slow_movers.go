package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/platform/logger"
)

// SlowMoverRow is a product with no sales in the requested window.
type SlowMoverRow struct {
	ProductID        int64
	Name             string
	StockQuantity    float64
	LastSoldAtExists bool
}

// GetSlowMovers returns active ERP products that have zero sales in the
// trailing `days` window. Products never sold are also included.
func (r *Repository) GetSlowMovers(ctx context.Context, businessID int64, days int32) ([]SlowMoverRow, error) {
	if days <= 0 {
		days = 60
	}
	if days > 365 {
		days = 365
	}
	cutoff := nowFn().AddDate(0, 0, -int(days))
	// cross-schema read from erp_products + pos_order_items + pos_orders — slow-mover detection
	const q = `
		SELECT p.id, p.name, p.stock_quantity,
		       EXISTS (
		         SELECT 1 FROM pos_order_items oi
		         JOIN pos_orders o ON o.id = oi.order_id AND o.business_id = p.business_id
		         WHERE oi.menu_item_id = p.id AND o.created_at >= ?
		       ) AS recent_sale
		FROM erp_products p
		WHERE p.business_id = ?
		  AND p.is_active = TRUE
		HAVING recent_sale = 0
		ORDER BY p.name ASC
		LIMIT 500
	`
	rows, err := r.pool.QueryContext(ctx, q, cutoff, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: slow movers query failed", logger.Error(err))
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Warn("insights: rows.Close failed", logger.Error(closeErr))
		}
	}()

	out := make([]SlowMoverRow, 0)
	for rows.Next() {
		var row SlowMoverRow
		var stock interface{}
		var recent int64
		if scanErr := rows.Scan(&row.ProductID, &row.Name, &stock, &recent); scanErr != nil {
			logger.Ctx(ctx).Error("insights: slow movers scan failed", logger.Error(scanErr))
			return nil, scanErr
		}
		row.StockQuantity = toFloat(stock)
		out = append(out, row)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	return out, nil
}
