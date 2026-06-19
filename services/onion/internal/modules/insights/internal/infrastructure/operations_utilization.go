package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/platform/logger"
)

// UtilizationRow is the rolled-up scheduled-vs-available hours snapshot.
type UtilizationRow struct {
	ScheduledHours float64
	AvailableHours float64
}

// GetUtilization estimates staff utilization as:
//
//	scheduled_hours / available_hours
//
// where scheduled_hours comes from the per-shift start/end time strings
// stored in hr_shifts joined to hr_shift_assignments.date in the next 7
// days, and available_hours is 40h/week times the active headcount.
//
// This is deliberately a rough heuristic at P3 — the strict calculation
// lives under hr_payroll and wouldn't fit in a dashboard card anyway.
func (r *Repository) GetUtilization(ctx context.Context, businessID int64) (UtilizationRow, error) {
	// cross-schema read from hr_employees — active headcount for the utilization denominator
	var active int64
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM hr_employees WHERE business_id = ? AND status = 'active'`,
		businessID,
	).Scan(&active); err != nil {
		logger.Ctx(ctx).Error("insights: utilization active count failed", logger.Error(err))
		return UtilizationRow{}, err
	}

	// cross-schema read from hr_shift_assignments + hr_shifts — scheduled hours for next 7 days
	const q = `
		SELECT COALESCE(SUM(
		  TIME_TO_SEC(STR_TO_DATE(s.end_time, '%H:%i'))
		  - TIME_TO_SEC(STR_TO_DATE(s.start_time, '%H:%i'))
		) / 3600, 0)
		FROM hr_shift_assignments a
		JOIN hr_shifts s ON s.id = a.shift_id AND s.business_id = a.business_id
		WHERE a.business_id = ?
		  AND a.status <> 'cancelled'
		  AND a.date BETWEEN CURDATE() AND DATE_ADD(CURDATE(), INTERVAL 7 DAY)
	`
	var scheduled interface{}
	if err := r.pool.QueryRowContext(ctx, q, businessID).Scan(&scheduled); err != nil {
		logger.Ctx(ctx).Error("insights: utilization scheduled failed", logger.Error(err))
		return UtilizationRow{}, err
	}

	return UtilizationRow{
		ScheduledHours: toFloat(scheduled),
		AvailableHours: float64(active) * 40.0, // 40h/wk standard
	}, nil
}
