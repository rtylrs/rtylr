package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type ProductKPIs = domain.ProductKPIs

func (r *Repository) GetProductKPIs(ctx context.Context, businessID int64) (ProductKPIs, error) {
	kpis, err := r.sources.ERP.InsightsProductKPIs(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: product kpis failed", logger.Error(err))
		return ProductKPIs{}, err
	}
	return ProductKPIs{
		TotalSKUs:    kpis.TotalSKUs,
		ActiveSKUs:   kpis.ActiveSKUs,
		OutOfStock:   kpis.OutOfStock,
		AvgMarginPct: kpis.AvgMarginPct,
	}, nil
}

// roundUnits keeps two decimal places, as product sales cards show them.
func roundUnits(units float64) float64 {
	return float64(decimalToMinorUnits(units)) / 100
}
