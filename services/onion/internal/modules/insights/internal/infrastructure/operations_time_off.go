package infrastructure

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type TimeOffSummaryRow = domain.TimeOffSummaryRow

func (r *Repository) GetTimeOffSummary(ctx context.Context, businessID int64, from, to time.Time) (TimeOffSummaryRow, error) {
	out := TimeOffSummaryRow{}
	var err error
	if out.Pending, err = r.sources.HR.InsightsLeaveStarting(ctx, businessID, "pending", dateOnly(from), dateOnly(to)); err != nil {
		logger.Ctx(ctx).Error("insights: pending time off failed", logger.Error(err))
		return TimeOffSummaryRow{}, err
	}
	if out.Approved, err = r.sources.HR.InsightsLeaveStarting(ctx, businessID, "approved", dateOnly(from), dateOnly(to)); err != nil {
		logger.Ctx(ctx).Error("insights: approved time off failed", logger.Error(err))
		return TimeOffSummaryRow{}, err
	}
	return out, nil
}
