package infrastructure

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type ShiftMetricsRow struct {
	Scheduled int64
	Completed int64
	NoShows   int64
}

func (r *Repository) GetShiftMetrics(ctx context.Context, businessID int64, from, to time.Time) (ShiftMetricsRow, error) {
	// cross-schema read from hr_shift_assignments — shifts scheduled/completed/no-show in window
	row, err := r.queries.GetShiftMetrics(ctx, sqlc.GetShiftMetricsParams{
		BusinessID: businessID,
		FromDate:   from,
		ToDate:     to,
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: shift metrics failed", logger.Error(err))
		return ShiftMetricsRow{}, err
	}
	return ShiftMetricsRow{
		Scheduled: ifaceToInt64(row.Scheduled),
		Completed: ifaceToInt64(row.Completed),
		NoShows:   ifaceToInt64(row.NoShows),
	}, nil
}
