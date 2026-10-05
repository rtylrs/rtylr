package infrastructure

import (
	"context"
	"github.com/voxire/onion/internal/modules/insights/internal/domain"

	"github.com/voxire/onion/internal/platform/logger"
)

type OperationsKPIs = domain.OperationsKPIs

func (r *Repository) GetOperationsKPIs(ctx context.Context, businessID int64) (OperationsKPIs, error) {
	out := OperationsKPIs{}

	// cross-schema read from hr_employees — active staff count for ops KPI card
	n, err := r.queries.CountActiveStaffOps(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: ops active staff failed", logger.Error(err))
		return OperationsKPIs{}, err
	}
	out.ActiveStaff = n

	// cross-schema read from hr_shift_assignments — shifts scheduled in the current week
	n, err = r.queries.CountShiftsThisWeek(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: ops shifts this week failed", logger.Error(err))
		return OperationsKPIs{}, err
	}
	out.ShiftsThisWeek = n

	// cross-schema read from hr_leave_requests — pending leave count for ops KPI card
	n, err = r.queries.CountPendingLeaveOps(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: ops pending leave failed", logger.Error(err))
		return OperationsKPIs{}, err
	}
	out.PendingTimeOff = n

	// cross-schema read from hr_leave_requests — any non-terminal leave rows
	n, err = r.queries.CountOpenLeaveRequests(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: ops open leave failed", logger.Error(err))
		return OperationsKPIs{}, err
	}
	out.OpenLeaveRequests = n

	return out, nil
}
