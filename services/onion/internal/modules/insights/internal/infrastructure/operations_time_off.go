package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/platform/logger"
)

// TimeOffSummaryRow is a simple two-counter summary of leave requests.
type TimeOffSummaryRow struct {
	Pending  int64
	Approved int64
}

// GetTimeOffSummary returns counts of pending and approved leave
// requests owned by the tenant.
func (r *Repository) GetTimeOffSummary(ctx context.Context, ownerID int64) (TimeOffSummaryRow, error) {
	out := TimeOffSummaryRow{}

	// cross-schema read from hr_leave_requests — pending leave requests
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM hr_leave_requests WHERE owner_id = ? AND status = 'pending'`,
		ownerID,
	).Scan(&out.Pending); err != nil {
		logger.Ctx(ctx).Error("insights: pending leave failed", logger.Error(err))
		return TimeOffSummaryRow{}, err
	}

	// cross-schema read from hr_leave_requests — approved leave requests still in effect
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM hr_leave_requests WHERE owner_id = ? AND status = 'approved'`,
		ownerID,
	).Scan(&out.Approved); err != nil {
		logger.Ctx(ctx).Error("insights: approved leave failed", logger.Error(err))
		return TimeOffSummaryRow{}, err
	}
	return out, nil
}
