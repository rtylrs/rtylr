package infrastructure

import (
	"context"
	"database/sql"
	"time"

	financecontract "github.com/voxire/onion/internal/contracts/finance"
	"github.com/voxire/onion/internal/modules/finance/finance_reports/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

// The reads below serve Insights reports. Each returns minor units keyed the
// way Insights charts them, so Insights never reads finance tables itself.

// InsightsPaidInvoicesByDay returns paid invoice totals keyed by reporting
// date (YYYY-MM-DD).
func (r *Repository) InsightsPaidInvoicesByDay(ctx context.Context, businessID int64, from, to time.Time) (map[string]int64, error) {
	rows, err := r.queries.InsightsPaidInvoicesByDay(ctx, sqlc.InsightsPaidInvoicesByDayParams{
		BusinessID: businessID,
		DateFrom:   sql.NullTime{Time: from, Valid: true},
		DateTo:     sql.NullTime{Time: to, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, row := range rows {
		if row.Day.Valid {
			out[row.Day.Time.Format(time.DateOnly)] = sqlutil.DecimalToMinorUnits(row.Amount)
		}
	}
	return out, nil
}

// InsightsExpensesByDay returns approved expense totals keyed by expense date.
func (r *Repository) InsightsExpensesByDay(ctx context.Context, businessID int64, from, to time.Time) (map[string]int64, error) {
	rows, err := r.queries.InsightsExpensesByDay(ctx, sqlc.InsightsExpensesByDayParams{BusinessID: businessID, DateFrom: from, DateTo: to})
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, row := range rows {
		out[row.Day.Format(time.DateOnly)] = sqlutil.DecimalToMinorUnits(row.Amount)
	}
	return out, nil
}

// InsightsExpensesByCategory returns approved expense totals per category,
// largest first.
func (r *Repository) InsightsExpensesByCategory(ctx context.Context, businessID int64, from, to time.Time) ([]financecontract.ExpenseCategoryTotal, error) {
	rows, err := r.queries.InsightsExpensesByCategory(ctx, sqlc.InsightsExpensesByCategoryParams{BusinessID: businessID, DateFrom: from, DateTo: to})
	if err != nil {
		return nil, err
	}
	out := make([]financecontract.ExpenseCategoryTotal, 0, len(rows))
	for _, row := range rows {
		out = append(out, financecontract.ExpenseCategoryTotal{
			CategoryID:       row.CatID,
			CategoryName:     row.CatName,
			AmountMinorUnits: sqlutil.DecimalToMinorUnits(row.Amount),
		})
	}
	return out, nil
}

// InsightsInvoiceAging returns outstanding invoices from the last 12 months
// grouped by due-date age. Buckets with no invoices are omitted.
func (r *Repository) InsightsInvoiceAging(ctx context.Context, businessID int64) ([]financecontract.InvoiceAgingBucket, error) {
	rows, err := r.queries.InsightsInvoiceAging(ctx, businessID)
	if err != nil {
		return nil, err
	}
	out := make([]financecontract.InvoiceAgingBucket, 0, len(rows))
	for _, row := range rows {
		out = append(out, financecontract.InvoiceAgingBucket{
			Bucket:           row.Bucket,
			InvoiceCount:     row.InvoiceCount,
			AmountMinorUnits: sqlutil.DecimalToMinorUnits(row.AmountDue),
		})
	}
	return out, nil
}

// InsightsInvoiceRevenue sums paid invoices by reporting date.
func (r *Repository) InsightsInvoiceRevenue(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	total, err := r.queries.SumInsightsInvoiceRevenue(ctx, sqlc.SumInsightsInvoiceRevenueParams{
		BusinessID: businessID,
		DateFrom:   sql.NullTime{Time: from, Valid: true},
		DateTo:     sql.NullTime{Time: to, Valid: true},
	})
	if err != nil {
		return 0, err
	}
	return sqlutil.DecimalToMinorUnits(total), nil
}

// InsightsLastChange returns the latest update to an invoice or expense.
func (r *Repository) InsightsLastChange(ctx context.Context, businessID int64) (time.Time, error) {
	return r.queries.InsightsFinanceLastChange(ctx, sqlc.InsightsFinanceLastChangeParams{BusinessID: businessID})
}
