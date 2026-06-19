package infrastructure

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

// HeadcountSnapshot carries the current active headcount alongside the
// value as-of one month ago so the service can produce a MoM delta.
type HeadcountSnapshot struct {
	ActiveHeadcount      int64
	HeadcountOneMonthAgo int64
}

// GetHeadcount returns active staff count plus a month-ago snapshot.
//
// "Active as-of date X" is approximated as: employees whose start_date is
// on or before X and whose end_date is NULL or after X. We cannot use
// status here because status is mutable; a terminated employee's
// historical active-state would be misrepresented by that column alone.
func (r *Repository) GetHeadcount(ctx context.Context, businessID int64) (HeadcountSnapshot, error) {
	now := nowFn()
	monthAgo := now.AddDate(0, -1, 0)

	var nowCount, priorCount int64

	// cross-schema read from hr_employees — current active headcount
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM hr_employees
		 WHERE business_id = ?
		   AND (start_date IS NULL OR start_date <= ?)
		   AND (end_date IS NULL OR end_date > ?)
		   AND status = 'active'`,
		businessID, now.Format("2006-01-02"), now.Format("2006-01-02"),
	).Scan(&nowCount); err != nil {
		logger.Ctx(ctx).Error("insights: headcount now failed", logger.Error(err))
		return HeadcountSnapshot{}, err
	}

	// cross-schema read from hr_employees — headcount as-of one month ago for MoM delta
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM hr_employees
		 WHERE business_id = ?
		   AND (start_date IS NULL OR start_date <= ?)
		   AND (end_date IS NULL OR end_date > ?)`,
		businessID, monthAgo.Format("2006-01-02"), monthAgo.Format("2006-01-02"),
	).Scan(&priorCount); err != nil {
		logger.Ctx(ctx).Error("insights: headcount prior failed", logger.Error(err))
		return HeadcountSnapshot{}, err
	}

	return HeadcountSnapshot{
		ActiveHeadcount:      nowCount,
		HeadcountOneMonthAgo: priorCount,
	}, nil
}

var _ = time.Time{} // keep `time` imported in case of future time param use
