package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/platform/logger"
)

type TimeOffSummaryRow struct {
	Pending  int64
	Approved int64
}

func (r *Repository) GetTimeOffSummary(ctx context.Context, businessID int64) (TimeOffSummaryRow, error) {
	out := TimeOffSummaryRow{}

	// cross-schema read from hr_leave_requests — pending leave requests
	n, err := r.queries.CountPendingTimeOff(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: pending leave failed", logger.Error(err))
		return TimeOffSummaryRow{}, err
	}
	out.Pending = n

	// cross-schema read from hr_leave_requests — approved leave requests still in effect
	n, err = r.queries.CountApprovedTimeOff(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: approved leave failed", logger.Error(err))
		return TimeOffSummaryRow{}, err
	}
	out.Approved = n
	return out, nil
}
