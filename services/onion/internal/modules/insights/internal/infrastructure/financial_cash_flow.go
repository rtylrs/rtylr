package infrastructure

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

// CashFlowDayRow is a single day in the cash-flow series. Inflow covers
// paid invoices; outflow covers logged expenses. Both in minor units.
type CashFlowDayRow struct {
	Date              string
	InflowMinorUnits  int64
	OutflowMinorUnits int64
}

// GetCashFlowSeries returns a contiguous daily series of (inflow, outflow)
// for the window. Missing days are filled with zero entries so the chart
// renders without gaps.
func (r *Repository) GetCashFlowSeries(ctx context.Context, ownerID int64, from, to time.Time) ([]CashFlowDayRow, error) {
	inflow, err := r.dailyPaidInvoices(ctx, ownerID, from, to)
	if err != nil {
		return nil, err
	}
	outflow, err := r.dailyExpenses(ctx, ownerID, from, to)
	if err != nil {
		return nil, err
	}

	out := make([]CashFlowDayRow, 0)
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		out = append(out, CashFlowDayRow{
			Date:              key,
			InflowMinorUnits:  inflow[key],
			OutflowMinorUnits: outflow[key],
		})
	}
	return out, nil
}

func (r *Repository) dailyPaidInvoices(ctx context.Context, ownerID int64, from, to time.Time) (map[string]int64, error) {
	// cross-schema read from finance_invoices — daily inflow for cash-flow chart
	const q = `
		SELECT DATE(COALESCE(paid_date, issue_date)) AS day, COALESCE(SUM(total), 0)
		FROM finance_invoices
		WHERE owner_id = ?
		  AND status = 'paid'
		  AND COALESCE(paid_date, issue_date) BETWEEN ? AND ?
		GROUP BY day
	`
	return r.scanDailySums(ctx, q, ownerID, from.Format("2006-01-02"), to.Format("2006-01-02"))
}

func (r *Repository) dailyExpenses(ctx context.Context, ownerID int64, from, to time.Time) (map[string]int64, error) {
	// cross-schema read from finance_expenses — daily outflow for cash-flow chart
	const q = `
		SELECT DATE(expense_date) AS day, COALESCE(SUM(amount), 0)
		FROM finance_expenses
		WHERE owner_id = ?
		  AND expense_date BETWEEN ? AND ?
		GROUP BY day
	`
	return r.scanDailySums(ctx, q, ownerID, from.Format("2006-01-02"), to.Format("2006-01-02"))
}

func (r *Repository) scanDailySums(ctx context.Context, q string, args ...interface{}) (map[string]int64, error) {
	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("insights: daily sum query failed", logger.Error(err))
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Warn("insights: rows.Close failed", logger.Error(closeErr))
		}
	}()

	out := make(map[string]int64)
	for rows.Next() {
		var day time.Time
		var amount interface{}
		if scanErr := rows.Scan(&day, &amount); scanErr != nil {
			logger.Ctx(ctx).Error("insights: daily sum scan failed", logger.Error(scanErr))
			return nil, scanErr
		}
		out[day.Format("2006-01-02")] = decimalToMinorUnits(amount)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	return out, nil
}
