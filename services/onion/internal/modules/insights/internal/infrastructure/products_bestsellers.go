package infrastructure

import (
	"context"
	"sort"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type BestSellerRow = domain.BestSellerRow

// GetBestSellers ranks the business's products by units sold on paid orders
// over the last 30 days.
func (r *Repository) GetBestSellers(ctx context.Context, businessID int64, limit int32) ([]BestSellerRow, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	cutoff := nowFn().AddDate(0, 0, -30)
	sales, err := r.sources.POS.InsightsPaidProductSales(ctx, businessID, cutoff, time.Time{})
	if err != nil {
		logger.Ctx(ctx).Error("insights: best sellers sales failed", logger.Error(err))
		return nil, err
	}
	facts, err := r.productFacts(ctx, businessID, sales)
	if err != nil {
		logger.Ctx(ctx).Error("insights: best sellers products failed", logger.Error(err))
		return nil, err
	}
	out := make([]BestSellerRow, 0, len(sales))
	for _, sale := range sales {
		product, ok := facts[sale.ProductID]
		if !ok {
			continue
		}
		out = append(out, BestSellerRow{
			ProductID: sale.ProductID, Name: product.Name,
			UnitsSold: roundUnits(sale.Units), RevenueMinorUnits: sale.RevenueMinorUnits,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].UnitsSold != out[j].UnitsSold {
			return out[i].UnitsSold > out[j].UnitsSold
		}
		return out[i].ProductID < out[j].ProductID
	})
	if len(out) > int(limit) {
		out = out[:limit]
	}
	return out, nil
}
