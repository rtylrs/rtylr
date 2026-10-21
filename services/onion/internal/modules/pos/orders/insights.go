package orders

import (
	"context"
	"time"

	poscontract "github.com/voxire/onion/internal/contracts/pos"
)

// Insights reports read order data through these methods rather than querying
// POS tables. Amounts are minor units unless stated otherwise.

func (m *Module) InsightsNetRevenue(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	return m.repository.InsightsNetRevenue(ctx, businessID, from, to)
}

func (m *Module) InsightsOrdersToday(ctx context.Context, businessID int64) (int64, float64, error) {
	return m.repository.InsightsOrdersToday(ctx, businessID)
}

func (m *Module) InsightsBuyerCount(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	return m.repository.InsightsBuyerCount(ctx, businessID, from, to)
}

func (m *Module) InsightsLapsedBuyerCount(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	return m.repository.InsightsLapsedBuyerCount(ctx, businessID, from, to)
}

func (m *Module) InsightsTopBuyers(ctx context.Context, businessID int64, query poscontract.TopBuyersQuery) ([]poscontract.CustomerPurchases, error) {
	return m.repository.InsightsTopBuyers(ctx, businessID, query)
}

func (m *Module) InsightsBuyerSpendBuckets(ctx context.Context, businessID int64) ([]poscontract.SpendBucket, error) {
	return m.repository.InsightsBuyerSpendBuckets(ctx, businessID)
}

func (m *Module) InsightsPaidProductSales(ctx context.Context, businessID int64, from, to time.Time) ([]poscontract.ProductSales, error) {
	return m.repository.InsightsPaidProductSales(ctx, businessID, from, to)
}

func (m *Module) InsightsLifetimeProductSales(ctx context.Context, businessID int64) ([]poscontract.ProductSales, error) {
	return m.repository.InsightsLifetimeProductSales(ctx, businessID)
}

func (m *Module) InsightsProductsSoldSince(ctx context.Context, businessID int64, since time.Time) ([]int64, error) {
	return m.repository.InsightsProductsSoldSince(ctx, businessID, since)
}

func (m *Module) InsightsLastChange(ctx context.Context, businessID int64) (time.Time, error) {
	return m.repository.InsightsLastChange(ctx, businessID)
}
