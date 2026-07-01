package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type HeadcountSnapshot struct {
	ActiveHeadcount      int64
	HeadcountOneMonthAgo int64
}

func dateParam(t time.Time) sql.NullTime {
	d := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return sql.NullTime{Time: d, Valid: true}
}

func (r *Repository) GetHeadcount(ctx context.Context, businessID int64) (HeadcountSnapshot, error) {
	now := nowFn()
	monthAgo := now.AddDate(0, -1, 0)

	// cross-schema read from hr_employees — current active headcount
	nowCount, err := r.queries.GetHeadcountNow(ctx, sqlc.GetHeadcountNowParams{
		BusinessID: businessID,
		StartDate:  dateParam(now),
		EndDate:    dateParam(now),
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: headcount now failed", logger.Error(err))
		return HeadcountSnapshot{}, err
	}

	// cross-schema read from hr_employees — headcount as-of one month ago for MoM delta
	priorCount, err := r.queries.GetHeadcountPrior(ctx, sqlc.GetHeadcountPriorParams{
		BusinessID: businessID,
		StartDate:  dateParam(monthAgo),
		EndDate:    dateParam(monthAgo),
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: headcount prior failed", logger.Error(err))
		return HeadcountSnapshot{}, err
	}

	return HeadcountSnapshot{ActiveHeadcount: nowCount, HeadcountOneMonthAgo: priorCount}, nil
}
