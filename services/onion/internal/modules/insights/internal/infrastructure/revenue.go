package infrastructure

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

func (r *Repository) GetRevenueSeries(ctx context.Context, businessID int64, from, to time.Time) ([]domain.RevenuePoint, error) {
	byDay, err := r.sources.Finance.InsightsPaidInvoicesByDay(ctx, businessID, from, to)
	if err != nil {
		logger.Ctx(ctx).Error("insights: revenue series query failed", logger.Error(err))
		return nil, err
	}
	series := make([]domain.RevenuePoint, 0)
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		series = append(series, domain.RevenuePoint{Date: key, AmountMinorUnits: byDay[key]})
	}
	return series, nil
}
