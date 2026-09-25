package infrastructure

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type UtilizationRow struct {
	ScheduledHours float64
	AvailableHours float64
}

func inclusiveWorkdays(from, to time.Time) int64 {
	start := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)
	if end.Before(start) {
		return 0
	}
	var count int64
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		if day.Weekday() != time.Saturday && day.Weekday() != time.Sunday {
			count++
		}
	}
	return count
}

func (r *Repository) GetUtilization(ctx context.Context, businessID int64, from, to time.Time) (UtilizationRow, error) {
	active, err := r.queries.CountEmployeesAsOf(ctx, sqlc.CountEmployeesAsOfParams{
		BusinessID: businessID,
		StartDate:  dateParam(to),
		EndDate:    dateParam(to),
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: utilization employee count failed", logger.Error(err))
		return UtilizationRow{}, err
	}

	scheduled, err := r.queries.SumScheduledHoursInRange(ctx, sqlc.SumScheduledHoursInRangeParams{
		BusinessID: businessID,
		Date:       dateOnly(from),
		Date_2:     dateOnly(to),
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: utilization scheduled hours failed", logger.Error(err))
		return UtilizationRow{}, err
	}

	return UtilizationRow{
		ScheduledHours: toFloat(scheduled),
		AvailableHours: float64(active) * float64(inclusiveWorkdays(from, to)) * 8,
	}, nil
}
