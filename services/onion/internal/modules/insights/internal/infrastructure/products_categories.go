package infrastructure

import (
	"context"
	"sort"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type CategorySalesRow = domain.CategorySalesRow

// GetSalesByCategory groups month-to-date paid sales by the product's ERP
// category. Products without a category are reported as Uncategorised.
func (r *Repository) GetSalesByCategory(ctx context.Context, businessID int64) ([]CategorySalesRow, error) {
	thisFrom, thisTo, _, _ := mtdRanges(nowFn())
	sales, err := r.sources.POS.InsightsPaidProductSales(ctx, businessID, thisFrom, thisTo)
	if err != nil {
		logger.Ctx(ctx).Error("insights: category sales failed", logger.Error(err))
		return nil, err
	}
	facts, err := r.productFacts(ctx, businessID, sales)
	if err != nil {
		logger.Ctx(ctx).Error("insights: category sales products failed", logger.Error(err))
		return nil, err
	}
	type total struct {
		name    string
		units   float64
		revenue int64
	}
	byCategory := make(map[int64]*total)
	for _, sale := range sales {
		product, ok := facts[sale.ProductID]
		if !ok {
			continue
		}
		category := byCategory[product.CategoryID]
		if category == nil {
			name := product.CategoryName
			if product.CategoryID == 0 {
				name = "Uncategorised"
			}
			category = &total{name: name}
			byCategory[product.CategoryID] = category
		}
		category.units += sale.Units
		category.revenue += sale.RevenueMinorUnits
	}
	out := make([]CategorySalesRow, 0, len(byCategory))
	for id, category := range byCategory {
		out = append(out, CategorySalesRow{
			CategoryID: id, CategoryName: category.name,
			UnitsSold: roundUnits(category.units), RevenueMinorUnits: category.revenue,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].RevenueMinorUnits != out[j].RevenueMinorUnits {
			return out[i].RevenueMinorUnits > out[j].RevenueMinorUnits
		}
		return out[i].CategoryID < out[j].CategoryID
	})
	return out, nil
}
