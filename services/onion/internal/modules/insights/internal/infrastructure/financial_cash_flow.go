package infrastructure

import (
	"context"
	"database/sql"
	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type CashFlowDayRow = domain.CashFlowDayRow

func (r *Repository) GetCashFlowSeries(ctx context.Context, businessID int64, from, to time.Time) ([]CashFlowDayRow, error) {
	inflow, err := r.dailyPaidInvoices(ctx, businessID, from, to)
	if err != nil {
		return nil, err
	}
	outflow, err := r.dailyExpenses(ctx, businessID, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]CashFlowDayRow, 0)
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		out = append(out, CashFlowDayRow{
			Date: key, InflowMinorUnits: inflow[key], OutflowMinorUnits: outflow[key],
		})
	}
	return out, nil
}

func (r *Repository) dailyPaidInvoices(ctx context.Context, businessID int64, from, to time.Time) (map[string]int64, error) {
	// cross-schema read from finance_invoices — daily inflow for cash-flow chart
	rows, err := r.queries.GetDailyPaidInvoices(ctx, sqlc.GetDailyPaidInvoicesParams{
		BusinessID:      businessID,
		ReportingDate:   sql.NullTime{Time: from, Valid: true},
		ReportingDate_2: sql.NullTime{Time: to, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: daily paid invoices failed", logger.Error(err))
		return nil, err
	}
	out := make(map[string]int64)
	for _, row := range rows {
		if row.Day.Valid {
			out[row.Day.Time.Format("2006-01-02")] = decimalToMinorUnits(row.Amount)
		}
	}
	return out, nil
}

func (r *Repository) dailyExpenses(ctx context.Context, businessID int64, from, to time.Time) (map[string]int64, error) {
	// cross-schema read from finance_expenses — daily outflow for cash-flow chart
	rows, err := r.queries.GetDailyExpenses(ctx, sqlc.GetDailyExpensesParams{
		BusinessID:      businessID,
		FromExpenseDate: from,
		ToExpenseDate:   to,
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: daily expenses failed", logger.Error(err))
		return nil, err
	}
	out := make(map[string]int64)
	for _, row := range rows {
		out[row.Day.Format("2006-01-02")] = decimalToMinorUnits(row.Amount)
	}
	return out, nil
}
