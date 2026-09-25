package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

func (r *Repository) GetRevenueSeries(ctx context.Context, businessID int64, from, to time.Time) ([]domain.RevenuePoint, error) {
	// cross-schema read from finance_invoices — daily paid revenue for the revenue line chart
	rows, err := r.queries.GetRevenueSeries(ctx, sqlc.GetRevenueSeriesParams{
		BusinessID:      businessID,
		ReportingDate:   sql.NullTime{Time: from, Valid: true},
		ReportingDate_2: sql.NullTime{Time: to, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: revenue series query failed", logger.Error(err))
		return nil, err
	}
	byDay := make(map[string]int64)
	for _, row := range rows {
		if row.Day.Valid {
			byDay[row.Day.Time.Format("2006-01-02")] = decimalToMinorUnits(row.Amount)
		}
	}
	series := make([]domain.RevenuePoint, 0)
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		series = append(series, domain.RevenuePoint{Date: key, AmountMinorUnits: byDay[key]})
	}
	return series, nil
}
