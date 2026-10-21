package infrastructure

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type HeadcountSnapshot = domain.HeadcountSnapshot

func (r *Repository) GetHeadcount(ctx context.Context, businessID int64, from, to time.Time) (HeadcountSnapshot, error) {
	periodStart := dateOnly(from)
	periodEnd := dateOnly(to)
	dayBeforeStart := dateOnly(from.AddDate(0, 0, -1))

	headcountAtStart, err := r.sources.HR.InsightsHeadcountAsOf(ctx, businessID, dayBeforeStart)
	if err != nil {
		logger.Ctx(ctx).Error("insights: headcount at start failed", logger.Error(err))
		return HeadcountSnapshot{}, err
	}
	activeHeadcount, err := r.sources.HR.InsightsHeadcountAsOf(ctx, businessID, periodEnd)
	if err != nil {
		logger.Ctx(ctx).Error("insights: active headcount failed", logger.Error(err))
		return HeadcountSnapshot{}, err
	}
	hired, err := r.sources.HR.InsightsHires(ctx, businessID, periodStart, periodEnd)
	if err != nil {
		logger.Ctx(ctx).Error("insights: hires count failed", logger.Error(err))
		return HeadcountSnapshot{}, err
	}
	departed, err := r.sources.HR.InsightsDepartures(ctx, businessID, periodStart, periodEnd)
	if err != nil {
		logger.Ctx(ctx).Error("insights: departures count failed", logger.Error(err))
		return HeadcountSnapshot{}, err
	}
	return HeadcountSnapshot{
		ActiveHeadcount:    activeHeadcount,
		HeadcountAtStart:   headcountAtStart,
		HiredThisPeriod:    hired,
		DepartedThisPeriod: departed,
	}, nil
}
