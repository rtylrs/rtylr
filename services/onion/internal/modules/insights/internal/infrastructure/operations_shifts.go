package infrastructure

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type ShiftMetricsRow = domain.ShiftMetricsRow

func (r *Repository) GetShiftMetrics(ctx context.Context, businessID int64, from, to time.Time) (ShiftMetricsRow, error) {
	metrics, err := r.sources.HR.InsightsShiftMetrics(ctx, businessID, from, to)
	if err != nil {
		logger.Ctx(ctx).Error("insights: shift metrics query failed", logger.Error(err))
		return ShiftMetricsRow{}, err
	}
	return ShiftMetricsRow{Scheduled: metrics.Scheduled, Completed: metrics.Completed, NoShows: metrics.NoShows}, nil
}
