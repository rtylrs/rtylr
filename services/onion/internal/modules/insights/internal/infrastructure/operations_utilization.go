package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/platform/logger"
)

type UtilizationRow struct {
	ScheduledHours float64
	AvailableHours float64
}

func (r *Repository) GetUtilization(ctx context.Context, businessID int64) (UtilizationRow, error) {
	// cross-schema read from hr_employees — active headcount for the utilization denominator
	active, err := r.queries.CountActiveEmployeesUtilization(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: utilization active count failed", logger.Error(err))
		return UtilizationRow{}, err
	}

	// cross-schema read from hr_shift_assignments + hr_shifts — scheduled hours for next 7 days
	scheduled, err := r.queries.SumScheduledHoursNext7Days(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: utilization scheduled failed", logger.Error(err))
		return UtilizationRow{}, err
	}

	return UtilizationRow{
		ScheduledHours: toFloat(scheduled),
		AvailableHours: float64(active) * 40.0,
	}, nil
}
