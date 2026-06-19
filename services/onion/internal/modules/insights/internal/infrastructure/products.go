package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/platform/logger"
)

// ProductKPIs summarises the state of the ERP product catalog.
type ProductKPIs struct {
	TotalSKUs    int64
	ActiveSKUs   int64
	OutOfStock   int64
	AvgMarginPct float64
}

// GetProductKPIs reads counts and an average margin from erp_products.
// Margin % per product = (sale_price - cost_price) / sale_price * 100
// when sale_price > 0; rows with zero sale_price are excluded from the
// average so they don't skew the result to negative infinity.
func (r *Repository) GetProductKPIs(ctx context.Context, businessID int64) (ProductKPIs, error) {
	out := ProductKPIs{}

	// cross-schema read from erp_products — total SKU count for the product KPI card
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM erp_products WHERE business_id = ?`,
		businessID,
	).Scan(&out.TotalSKUs); err != nil {
		logger.Ctx(ctx).Error("insights: product total failed", logger.Error(err))
		return ProductKPIs{}, err
	}

	// cross-schema read from erp_products — active SKU count
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM erp_products WHERE business_id = ? AND is_active = TRUE`,
		businessID,
	).Scan(&out.ActiveSKUs); err != nil {
		logger.Ctx(ctx).Error("insights: product active failed", logger.Error(err))
		return ProductKPIs{}, err
	}

	// cross-schema read from erp_products — out-of-stock count
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM erp_products
		 WHERE business_id = ?
		   AND track_inventory = TRUE
		   AND stock_quantity <= 0`,
		businessID,
	).Scan(&out.OutOfStock); err != nil {
		logger.Ctx(ctx).Error("insights: product oos failed", logger.Error(err))
		return ProductKPIs{}, err
	}

	// cross-schema read from erp_products — average margin across priced, sellable SKUs
	var margin interface{}
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COALESCE(AVG((sale_price - cost_price) / sale_price * 100), 0)
		 FROM erp_products
		 WHERE business_id = ?
		   AND is_active = TRUE
		   AND sale_price > 0`,
		businessID,
	).Scan(&margin); err != nil {
		logger.Ctx(ctx).Error("insights: product margin failed", logger.Error(err))
		return ProductKPIs{}, err
	}
	out.AvgMarginPct = toFloat(margin)
	return out, nil
}

// toFloat is a local mirror of decimalToMinorUnits for percentage values,
// which we want to keep as float64 rather than converting to minor units.
func toFloat(v interface{}) float64 {
	cents := decimalToMinorUnits(v)
	return float64(cents) / 100
}
