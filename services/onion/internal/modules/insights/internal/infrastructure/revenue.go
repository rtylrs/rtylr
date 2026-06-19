package infrastructure

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// GetRevenueSeries returns a daily revenue series (paid invoices only) for
// the [from, to] window, inclusive. Missing days are represented as zero
// entries so the frontend charting library can render a contiguous axis.
func (r *Repository) GetRevenueSeries(ctx context.Context, businessID int64, from, to time.Time) ([]domain.RevenuePoint, error) {
	// cross-schema read from finance_invoices — daily paid revenue for the revenue line chart
	const q = `
		SELECT DATE(COALESCE(paid_date, issue_date)) AS day,
		       COALESCE(SUM(total), 0)               AS amount
		FROM finance_invoices
		WHERE business_id = ?
		  AND status = 'paid'
		  AND COALESCE(paid_date, issue_date) >= ?
		  AND COALESCE(paid_date, issue_date) <= ?
		GROUP BY day
		ORDER BY day ASC
	`
	rows, err := r.pool.QueryContext(ctx, q, businessID, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		logger.Ctx(ctx).Error("insights: revenue series query failed", logger.Error(err))
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Warn("insights: rows.Close failed", logger.Error(closeErr))
		}
	}()

	byDay := make(map[string]int64)
	for rows.Next() {
		var day time.Time
		var amount interface{}
		if scanErr := rows.Scan(&day, &amount); scanErr != nil {
			logger.Ctx(ctx).Error("insights: revenue series scan failed", logger.Error(scanErr))
			return nil, scanErr
		}
		byDay[day.Format("2006-01-02")] = decimalToMinorUnits(amount)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}

	// Backfill zero-days to keep the series contiguous.
	days := int(to.Sub(from).Hours()/24) + 1
	if days < 0 {
		days = 0
	}
	series := make([]domain.RevenuePoint, 0, days)
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		series = append(series, domain.RevenuePoint{
			Date:             key,
			AmountMinorUnits: byDay[key],
		})
	}
	return series, nil
}
