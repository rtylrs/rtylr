package infrastructure

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type HeadcountSnapshot struct {
	ActiveHeadcount    int64
	HeadcountAtStart   int64
	HiredThisPeriod    int64
	DepartedThisPeriod int64
}

func (r *Repository) GetHeadcount(ctx context.Context, businessID int64, from, to time.Time) (HeadcountSnapshot, error) {
	periodStart := dateParam(from)
	periodEnd := dateParam(to)
	dayBeforeStart := dateParam(from.AddDate(0, 0, -1))

	headcountAtStart, err := r.queries.GetHeadcountAsOf(ctx, sqlc.GetHeadcountAsOfParams{
		BusinessID: businessID,
		StartDate:  dayBeforeStart,
		EndDate:    dayBeforeStart,
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: headcount period start failed", logger.Error(err))
		return HeadcountSnapshot{}, err
	}

	activeHeadcount, err := r.queries.GetHeadcountAsOf(ctx, sqlc.GetHeadcountAsOfParams{
		BusinessID: businessID,
		StartDate:  periodEnd,
		EndDate:    periodEnd,
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: headcount period end failed", logger.Error(err))
		return HeadcountSnapshot{}, err
	}

	hired, err := r.queries.CountHeadcountHires(ctx, sqlc.CountHeadcountHiresParams{
		BusinessID:  businessID,
		StartDate:   periodStart,
		StartDate_2: periodEnd,
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: headcount hires failed", logger.Error(err))
		return HeadcountSnapshot{}, err
	}

	departed, err := r.queries.CountHeadcountDepartures(ctx, sqlc.CountHeadcountDeparturesParams{
		BusinessID: businessID,
		EndDate:    periodStart,
		EndDate_2:  periodEnd,
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: headcount departures failed", logger.Error(err))
		return HeadcountSnapshot{}, err
	}

	return HeadcountSnapshot{
		ActiveHeadcount:    activeHeadcount,
		HeadcountAtStart:   headcountAtStart,
		HiredThisPeriod:    hired,
		DepartedThisPeriod: departed,
	}, nil
}
