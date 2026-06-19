package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/platform/logger"
)

// OperationsKPIs is the headline HR-ops snapshot returned at /insights/operations.
// Every counter here is derived from HR schemas via read-only aggregation
// (HR apps are off-limits at the app layer, but this read path is explicitly
// in-scope per the task spec).
type OperationsKPIs struct {
	ActiveStaff       int64
	ShiftsThisWeek    int64
	PendingTimeOff    int64
	OpenLeaveRequests int64
}

// GetOperationsKPIs aggregates the four headline ops counters.
func (r *Repository) GetOperationsKPIs(ctx context.Context, businessID int64) (OperationsKPIs, error) {
	out := OperationsKPIs{}

	// cross-schema read from hr_employees — active staff count for ops KPI card
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM hr_employees WHERE business_id = ? AND status = 'active'`,
		businessID,
	).Scan(&out.ActiveStaff); err != nil {
		logger.Ctx(ctx).Error("insights: ops active staff failed", logger.Error(err))
		return OperationsKPIs{}, err
	}

	// cross-schema read from hr_shift_assignments — shifts scheduled in the current week
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*)
		 FROM hr_shift_assignments
		 WHERE business_id = ?
		   AND date >= DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY)
		   AND date <  DATE_ADD(DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY), INTERVAL 7 DAY)`,
		businessID,
	).Scan(&out.ShiftsThisWeek); err != nil {
		logger.Ctx(ctx).Error("insights: ops shifts this week failed", logger.Error(err))
		return OperationsKPIs{}, err
	}

	// cross-schema read from hr_leave_requests — pending leave count for ops KPI card
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM hr_leave_requests WHERE business_id = ? AND status = 'pending'`,
		businessID,
	).Scan(&out.PendingTimeOff); err != nil {
		logger.Ctx(ctx).Error("insights: ops pending leave failed", logger.Error(err))
		return OperationsKPIs{}, err
	}

	// cross-schema read from hr_leave_requests — any non-terminal (pending/approved/in-review) leave rows
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM hr_leave_requests
		 WHERE business_id = ?
		   AND status NOT IN ('rejected','cancelled','completed')`,
		businessID,
	).Scan(&out.OpenLeaveRequests); err != nil {
		logger.Ctx(ctx).Error("insights: ops open leave failed", logger.Error(err))
		return OperationsKPIs{}, err
	}

	return out, nil
}
