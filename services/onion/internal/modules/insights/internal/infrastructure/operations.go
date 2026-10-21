package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type OperationsKPIs = domain.OperationsKPIs

func (r *Repository) GetOperationsKPIs(ctx context.Context, businessID int64) (OperationsKPIs, error) {
	out := OperationsKPIs{}
	var err error
	if out.ActiveStaff, err = r.sources.HR.InsightsActiveEmployees(ctx, businessID); err != nil {
		logger.Ctx(ctx).Error("insights: active staff count failed", logger.Error(err))
		return OperationsKPIs{}, err
	}
	if out.ShiftsThisWeek, err = r.sources.HR.InsightsShiftsThisWeek(ctx, businessID); err != nil {
		logger.Ctx(ctx).Error("insights: shifts this week failed", logger.Error(err))
		return OperationsKPIs{}, err
	}
	if out.PendingTimeOff, err = r.sources.HR.InsightsPendingLeave(ctx, businessID); err != nil {
		logger.Ctx(ctx).Error("insights: pending leave count failed", logger.Error(err))
		return OperationsKPIs{}, err
	}
	if out.OpenLeaveRequests, err = r.sources.HR.InsightsOpenLeave(ctx, businessID); err != nil {
		logger.Ctx(ctx).Error("insights: open leave count failed", logger.Error(err))
		return OperationsKPIs{}, err
	}
	return out, nil
}
