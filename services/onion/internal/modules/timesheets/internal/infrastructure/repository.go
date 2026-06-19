package infrastructure

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/timesheets/internal/domain"
	"github.com/voxire/onion/internal/modules/timesheets/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) GetTimesheets(ctx context.Context, businessID int64, weekStart string) ([]domain.TimesheetEntry, error) {
	// Dynamic WHERE clause and complex aggregation - built with squirrel
	builder := sq.
		Select(
			"s.staff_id",
			"CONCAT(st.first_name, ' ', st.last_name) AS staff_name",
			"DATE(s.date - INTERVAL WEEKDAY(s.date) DAY) AS week_start",
			"COALESCE(SUM(s.hours_worked), 0) AS total_hours",
			"COALESCE(SUM(s.hours_worked) - SUM(s.overtime_hours), 0) AS regular_hours",
			"COALESCE(SUM(s.overtime_hours), 0) AS overtime_hours",
		).
		From("hr_shifts s").
		Join("hr_staff st ON st.id = s.staff_id AND st.business_id = s.business_id").
		Where(sq.Eq{"s.business_id": businessID}).
		Where(sq.Eq{"s.status": "completed"})

	if weekStart != "" {
		builder = builder.Where(sq.Expr("DATE(s.date - INTERVAL WEEKDAY(s.date) DAY) = ?", weekStart))
	}

	builder = builder.
		GroupBy("s.staff_id", "staff_name", "week_start").
		OrderBy("week_start DESC", "staff_name ASC")

	query, args, err := builder.PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build timesheets query", logger.Error(err))
		return nil, err
	}

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
