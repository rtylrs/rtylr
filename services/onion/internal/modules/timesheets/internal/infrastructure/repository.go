package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/timesheets/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) GetTimesheets(ctx context.Context, ownerID int64, weekStart string) ([]domain.TimesheetEntry, error) {
	query := `
		SELECT s.staff_id,
		       CONCAT(st.first_name, ' ', st.last_name) AS staff_name,
		       DATE(s.date - INTERVAL WEEKDAY(s.date) DAY) AS week_start,
		       COALESCE(SUM(s.hours_worked), 0) AS total_hours,
		       COALESCE(SUM(s.hours_worked) - SUM(s.overtime_hours), 0) AS regular_hours,
		       COALESCE(SUM(s.overtime_hours), 0) AS overtime_hours
		FROM hr_shifts s
		JOIN hr_staff st ON st.id = s.staff_id AND st.owner_id = s.owner_id
		WHERE s.owner_id = ? AND s.status = 'completed'`

	args := []interface{}{ownerID}

	if weekStart != "" {
		query += ` AND DATE(s.date - INTERVAL WEEKDAY(s.date) DAY) = ?`
		args = append(args, weekStart)
	}

	query += ` GROUP BY s.staff_id, staff_name, week_start ORDER BY week_start DESC, staff_name ASC`

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get timesheets", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var entries []domain.TimesheetEntry
	for rows.Next() {
		var e domain.TimesheetEntry
		if err := rows.Scan(&e.StaffID, &e.StaffName, &e.WeekStart, &e.TotalHours, &e.RegularHours, &e.OvertimeHours); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
