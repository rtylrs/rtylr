package products

import (
	"context"
	"time"

	erpcontract "github.com/voxire/onion/internal/contracts/erp"
)

// Insights reports read catalogue data through these methods rather than
// querying ERP tables.

func (m *Module) InsightsProductKPIs(ctx context.Context, businessID int64) (erpcontract.ProductKPIs, error) {
	return m.repository.InsightsProductKPIs(ctx, businessID)
}

func (m *Module) InsightsLowStockCount(ctx context.Context, businessID int64) (int64, error) {
	return m.repository.InsightsLowStockCount(ctx, businessID)
}

func (m *Module) InsightsMarginDistribution(ctx context.Context, businessID int64) ([]int64, error) {
	return m.repository.InsightsMarginDistribution(ctx, businessID)
}

func (m *Module) InsightsProductFacts(ctx context.Context, businessID int64, ids []int64) ([]erpcontract.ProductFacts, error) {
	return m.repository.InsightsProductFacts(ctx, businessID, ids)
}

func (m *Module) InsightsActiveProducts(ctx context.Context, businessID int64) ([]erpcontract.StockedProduct, error) {
	return m.repository.InsightsActiveProducts(ctx, businessID)
}

func (m *Module) InsightsLastChange(ctx context.Context, businessID int64) (time.Time, error) {
	return m.repository.InsightsLastChange(ctx, businessID)
}
