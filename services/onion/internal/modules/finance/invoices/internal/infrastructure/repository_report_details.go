package infrastructure

import (
	"context"
	"sort"
	"time"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

const isoDate = "2006-01-02"

// SumPendingExpenses totals the expenses dated in [dateFrom, dateTo] that are
// still awaiting approval.
func (r *Repository) SumPendingExpenses(ctx context.Context, businessID int64, dateFrom, dateTo string) (float64, int64, error) {
	from, to, err := parseRange(dateFrom, dateTo)
	if err != nil {
		return 0, 0, err
	}
	row, err := r.queries.SumPendingExpenses(ctx, sqlc.SumPendingExpensesParams{
		BusinessID:      businessID,
		FromExpenseDate: from,
		ToExpenseDate:   to,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum pending expenses", logger.Error(err))
		return 0, 0, err
	}
	return toFloat64FromInterface(row.Total), row.ExpenseCount, nil
}

// SumExpensesByCategory splits the approved expenses dated in
// [dateFrom, dateTo] by category, largest first.
func (r *Repository) SumExpensesByCategory(ctx context.Context, businessID int64, dateFrom, dateTo string) ([]domain.CategoryTotal, error) {
	from, to, err := parseRange(dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.SumExpensesByCategory(ctx, sqlc.SumExpensesByCategoryParams{
		BusinessID:      businessID,
		FromExpenseDate: from,
		ToExpenseDate:   to,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum expenses by category", logger.Error(err))
		return nil, err
	}
	out := make([]domain.CategoryTotal, 0, len(rows))
	for _, row := range rows {
		total := domain.CategoryTotal{Name: row.Name, Color: row.Color, Amount: toFloat64FromInterface(row.Total)}
		if row.CategoryID.Valid {
			id := row.CategoryID.Int64
			total.CategoryID = &id
		}
		out = append(out, total)
	}
	return out, nil
}

// MonthlyTotals returns revenue and approved expenses for every calendar month
// from the month of dateFrom to the month of dateTo, months without activity
// included as zeros.
func (r *Repository) MonthlyTotals(ctx context.Context, businessID int64, dateFrom, dateTo string) ([]domain.MonthTotal, error) {
	from, to, err := parseRange(dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	revenue, err := r.queries.MonthlyPaidInvoiceTotals(ctx, sqlc.MonthlyPaidInvoiceTotalsParams{
		BusinessID:   businessID,
		FromPaidDate: sqlNullTimeStr(dateFrom),
		ToPaidDate:   sqlNullTimeStr(dateTo),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to total monthly revenue", logger.Error(err))
		return nil, err
	}
	expenses, err := r.queries.MonthlyExpenseTotals(ctx, sqlc.MonthlyExpenseTotalsParams{
		BusinessID:      businessID,
		FromExpenseDate: from,
		ToExpenseDate:   to,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to total monthly expenses", logger.Error(err))
		return nil, err
	}

	byMonth := map[string]*domain.MonthTotal{}
	var months []domain.MonthTotal
	for m := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, time.UTC); !m.After(to); m = m.AddDate(0, 1, 0) {
		months = append(months, domain.MonthTotal{Month: m.Format("2006-01")})
	}
	for i := range months {
		byMonth[months[i].Month] = &months[i]
	}
	for _, row := range revenue {
		if m, ok := byMonth[row.Month]; ok {
			m.Revenue = toFloat64FromInterface(row.Total)
		}
	}
	for _, row := range expenses {
		if m, ok := byMonth[row.Month]; ok {
			m.Expenses = toFloat64FromInterface(row.Total)
		}
	}
	return months, nil
}

// RecentActivity merges the latest expenses and paid invoices, newest first.
func (r *Repository) RecentActivity(ctx context.Context, businessID int64, limit int32) ([]domain.Activity, error) {
	expenses, err := r.queries.ListRecentExpenses(ctx, sqlc.ListRecentExpensesParams{BusinessID: businessID, Limit: limit})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list recent expenses", logger.Error(err))
		return nil, err
	}
	paid, err := r.queries.ListRecentPaidInvoices(ctx, sqlc.ListRecentPaidInvoicesParams{BusinessID: businessID, Limit: limit})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list recent paid invoices", logger.Error(err))
		return nil, err
	}

	out := make([]domain.Activity, 0, len(expenses)+len(paid))
	for _, e := range expenses {
		out = append(out, domain.Activity{
			Kind:        "expense",
			ID:          e.ID,
			Description: e.Description,
			Amount:      parseFloat(e.Amount),
			Date:        e.ExpenseDate.Format(isoDate),
			Status:      e.Status.String,
		})
	}
	for _, inv := range paid {
		description := inv.InvoiceNumber
		if inv.ClientName != "" {
			description += " · " + inv.ClientName
		}
		out = append(out, domain.Activity{
			Kind:        "payment",
			ID:          inv.ID,
			Description: description,
			Amount:      parseFloatNull(inv.Total),
			Date:        inv.PaidDate.Time.Format(isoDate),
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Date > out[j].Date })
	if len(out) > int(limit) {
		out = out[:limit]
	}
	return out, nil
}

func parseRange(dateFrom, dateTo string) (time.Time, time.Time, error) {
	from, err := time.Parse(isoDate, dateFrom)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	to, err := time.Parse(isoDate, dateTo)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return from, to, nil
}
