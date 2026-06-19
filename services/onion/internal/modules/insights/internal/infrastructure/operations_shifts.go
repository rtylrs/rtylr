package infrastructure

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

// ShiftMetricsRow carries scheduled vs completed shift counts over a window.
type ShiftMetricsRow struct {
	Scheduled int64
	Completed int64
	NoShows   int64
}

// GetShiftMetrics returns shift scheduling vs completion counts for the
// window, by reading hr_shift_assignments.status. "Scheduled" is the
// total of non-cancelled rows; "Completed" matches status = 'completed'.
// No-shows are explicit: status = 'no_show'.
func (r *Repository) GetShiftMetrics(ctx context.Context, businessID int64, from, to time.Time) (ShiftMetricsRow, error) {
	// cross-schema read from hr_shift_assignments — shifts scheduled/completed/no-show in window
	const q = `
		SELECT
		  SUM(CASE WHEN status <> 'cancelled' THEN 1 ELSE 0 END) AS scheduled,
		  SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END)  AS completed,
		  SUM(CASE WHEN status = 'no_show'   THEN 1 ELSE 0 END)  AS no_shows
		FROM hr_shift_assignments
		WHERE business_id = ?
		  AND date BETWEEN ? AND ?
	`
	var scheduled, completed, noShows int64
	if err := r.pool.QueryRowContext(ctx, q, businessID, from.Format("2006-01-02"), to.Format("2006-01-02")).Scan(
		&nullableInt64{&scheduled}, &nullableInt64{&completed}, &nullableInt64{&noShows},
	); err != nil {
		logger.Ctx(ctx).Error("insights: shift metrics failed", logger.Error(err))
		return ShiftMetricsRow{}, err
	}
	return ShiftMetricsRow{Scheduled: scheduled, Completed: completed, NoShows: noShows}, nil
}
