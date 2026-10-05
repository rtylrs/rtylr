package infrastructure

import (
	"context"
	"github.com/voxire/onion/internal/modules/insights/internal/domain"

	"github.com/voxire/onion/internal/platform/logger"
)

type ProductKPIs = domain.ProductKPIs

func (r *Repository) GetProductKPIs(ctx context.Context, businessID int64) (ProductKPIs, error) {
	out := ProductKPIs{}

	// cross-schema read from erp_products — total SKU count for the product KPI card
	n, err := r.queries.CountProductSKUsTotal(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: product total failed", logger.Error(err))
		return ProductKPIs{}, err
	}
	out.TotalSKUs = n

	// cross-schema read from erp_products — active SKU count
	n, err = r.queries.CountProductSKUsActive(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: product active failed", logger.Error(err))
		return ProductKPIs{}, err
	}
	out.ActiveSKUs = n

	// cross-schema read from erp_products — out-of-stock count
	n, err = r.queries.CountProductSKUsOutOfStock(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: product oos failed", logger.Error(err))
		return ProductKPIs{}, err
	}
	out.OutOfStock = n

	// cross-schema read from erp_products — average margin across priced, sellable SKUs
	margin, err := r.queries.AvgProductMarginPct(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: product margin failed", logger.Error(err))
		return ProductKPIs{}, err
	}
	out.AvgMarginPct = toFloat(margin)
	return out, nil
}

func toFloat(v interface{}) float64 {
	cents := decimalToMinorUnits(v)
	return float64(cents) / 100
}
