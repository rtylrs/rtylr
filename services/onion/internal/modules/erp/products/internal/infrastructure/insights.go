package infrastructure

import (
	"context"
	"time"

	erpcontract "github.com/voxire/onion/internal/contracts/erp"
	"github.com/voxire/onion/internal/modules/erp/products/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

// The reads below serve Insights reports, so Insights never queries ERP
// tables itself.

func (r *Repository) InsightsProductKPIs(ctx context.Context, businessID int64) (erpcontract.ProductKPIs, error) {
	var out erpcontract.ProductKPIs
	var err error
	if out.TotalSKUs, err = r.queries.InsightsCountProducts(ctx, businessID); err != nil {
		return erpcontract.ProductKPIs{}, err
	}
	if out.ActiveSKUs, err = r.queries.InsightsCountActiveProducts(ctx, businessID); err != nil {
		return erpcontract.ProductKPIs{}, err
	}
	if out.OutOfStock, err = r.queries.InsightsCountOutOfStock(ctx, businessID); err != nil {
		return erpcontract.ProductKPIs{}, err
	}
	margin, err := r.queries.InsightsAvgMarginPct(ctx, businessID)
	if err != nil {
		return erpcontract.ProductKPIs{}, err
	}
	out.AvgMarginPct = float64(sqlutil.DecimalToMinorUnits(margin)) / 100
	return out, nil
}

func (r *Repository) InsightsLowStockCount(ctx context.Context, businessID int64) (int64, error) {
	return r.queries.InsightsCountLowStock(ctx, businessID)
}

// InsightsMarginDistribution counts active priced products per margin band:
// below 0%, 0-10%, 10-25%, 25-50%, 50-75% and 75% or more.
func (r *Repository) InsightsMarginDistribution(ctx context.Context, businessID int64) ([]int64, error) {
	row, err := r.queries.InsightsMarginDistribution(ctx, businessID)
	if err != nil {
		return nil, err
	}
	return []int64{
		row.BelowZeroCount,
		row.ZeroToTenCount,
		row.TenToTwentyFiveCount,
		row.TwentyFiveToFiftyCount,
		row.FiftyToSeventyFiveCount,
		row.SeventyFivePlusCount,
	}, nil
}

// InsightsProductFacts returns the business's products among ids. IDs that
// belong to another business or no longer exist are left out.
func (r *Repository) InsightsProductFacts(ctx context.Context, businessID int64, ids []int64) ([]erpcontract.ProductFacts, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.queries.InsightsProductFacts(ctx, sqlc.InsightsProductFactsParams{BusinessID: businessID, Ids: ids})
	if err != nil {
		return nil, err
	}
	out := make([]erpcontract.ProductFacts, 0, len(rows))
	for _, row := range rows {
		out = append(out, erpcontract.ProductFacts{
			ID:             row.ID,
			Name:           row.Name,
			CategoryID:     row.CategoryID,
			CategoryName:   row.CategoryName,
			CostMinorUnits: row.CostMinorUnits,
		})
	}
	return out, nil
}

// InsightsActiveProducts returns every active product ordered by name.
func (r *Repository) InsightsActiveProducts(ctx context.Context, businessID int64) ([]erpcontract.StockedProduct, error) {
	rows, err := r.queries.InsightsActiveProducts(ctx, businessID)
	if err != nil {
		return nil, err
	}
	out := make([]erpcontract.StockedProduct, 0, len(rows))
	for _, row := range rows {
		out = append(out, erpcontract.StockedProduct{
			ID:            row.ID,
			Name:          row.Name,
			StockQuantity: float64(sqlutil.DecimalToMinorUnits(row.StockQuantity)) / 100,
		})
	}
	return out, nil
}

func (r *Repository) InsightsLastChange(ctx context.Context, businessID int64) (time.Time, error) {
	return r.queries.InsightsERPLastChange(ctx, businessID)
}
