package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/platform/logger"
)

// CategorySalesRow is one category of ERP products with its MTD revenue
// and units-sold totals rolled up from POS.
type CategorySalesRow struct {
	CategoryID        int64
	CategoryName      string
	UnitsSold         float64
	RevenueMinorUnits int64
}

// GetSalesByCategory returns MTD sales grouped by erp_categories.name.
// Products without a category are bundled under "Uncategorised" so the
// donut chart never has an orphan segment.
func (r *Repository) GetSalesByCategory(ctx context.Context, businessID int64) ([]CategorySalesRow, error) {
	thisFrom, thisTo, _, _ := mtdRanges(nowFn())
	// cross-schema read from erp_categories + erp_products + pos_order_items + pos_orders — MTD category sales
	const q = `
		SELECT COALESCE(c.id, 0) AS category_id,
		       COALESCE(c.name, 'Uncategorised') AS category_name,
		       COALESCE(SUM(oi.quantity), 0) AS units,
		       COALESCE(SUM(oi.total), 0)    AS revenue
		FROM pos_order_items oi
		JOIN pos_orders o        ON o.id = oi.order_id
		JOIN erp_products p      ON p.id = oi.menu_item_id AND p.business_id = o.business_id
		LEFT JOIN erp_categories c ON c.id = p.category_id AND c.business_id = p.business_id
		WHERE o.business_id = ?
		  AND o.created_at >= ?
		  AND o.created_at <= ?
		GROUP BY c.id, c.name
		ORDER BY revenue DESC
	`
	rows, err := r.pool.QueryContext(ctx, q, businessID, thisFrom, thisTo)
	if err != nil {
		logger.Ctx(ctx).Error("insights: category sales query failed", logger.Error(err))
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Warn("insights: rows.Close failed", logger.Error(closeErr))
		}
	}()

	out := make([]CategorySalesRow, 0)
	for rows.Next() {
		var row CategorySalesRow
		var units, revenue interface{}
		if scanErr := rows.Scan(&row.CategoryID, &row.CategoryName, &units, &revenue); scanErr != nil {
			logger.Ctx(ctx).Error("insights: category sales scan failed", logger.Error(scanErr))
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
