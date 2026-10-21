package financereports

import (
	"context"
	"time"

	financecontract "github.com/voxire/onion/internal/contracts/finance"
)

// Insights reports read finance data through these methods rather than
// querying finance tables. Amounts are minor units.

func (m *Module) InsightsPaidInvoicesByDay(ctx context.Context, businessID int64, from, to time.Time) (map[string]int64, error) {
	return m.repository.InsightsPaidInvoicesByDay(ctx, businessID, from, to)
}

func (m *Module) InsightsExpensesByDay(ctx context.Context, businessID int64, from, to time.Time) (map[string]int64, error) {
	return m.repository.InsightsExpensesByDay(ctx, businessID, from, to)
}

func (m *Module) InsightsExpensesByCategory(ctx context.Context, businessID int64, from, to time.Time) ([]financecontract.ExpenseCategoryTotal, error) {
	return m.repository.InsightsExpensesByCategory(ctx, businessID, from, to)
}

func (m *Module) InsightsInvoiceAging(ctx context.Context, businessID int64) ([]financecontract.InvoiceAgingBucket, error) {
	return m.repository.InsightsInvoiceAging(ctx, businessID)
}

func (m *Module) InsightsInvoiceRevenue(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	return m.repository.InsightsInvoiceRevenue(ctx, businessID, from, to)
}

func (m *Module) InsightsLastChange(ctx context.Context, businessID int64) (time.Time, error) {
	return m.repository.InsightsLastChange(ctx, businessID)
}
