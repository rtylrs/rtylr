package infrastructure

import (
	"context"
	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type TimeOffSummaryRow = domain.TimeOffSummaryRow

func (r *Repository) GetTimeOffSummary(ctx context.Context, businessID int64, from, to time.Time) (TimeOffSummaryRow, error) {
	out := TimeOffSummaryRow{}

	pending, err := r.queries.CountPendingTimeOff(ctx, sqlc.CountPendingTimeOffParams{
		BusinessID:    businessID,
		FromStartDate: dateOnly(from),
		ToStartDate:   dateOnly(to),
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: pending leave failed", logger.Error(err))
		return TimeOffSummaryRow{}, err
	}
	out.Pending = pending

	approved, err := r.queries.CountApprovedTimeOff(ctx, sqlc.CountApprovedTimeOffParams{
		BusinessID:    businessID,
		FromStartDate: dateOnly(from),
		ToStartDate:   dateOnly(to),
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: approved leave failed", logger.Error(err))
		return TimeOffSummaryRow{}, err
	}
	out.Approved = approved
	return out, nil
}
