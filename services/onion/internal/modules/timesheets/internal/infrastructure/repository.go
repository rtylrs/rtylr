package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/timesheets/internal/domain"
	"github.com/voxire/onion/internal/modules/timesheets/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

var errWeekStartNotMonday = errors.New("week start must be a Monday")

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

type timesheetQueryRow struct {
	staffID       sql.NullInt64
	staffName     string
	weekStart     time.Time
	totalHours    any
	regularHours  any
	overtimeHours any
}

func NewRepository(db *sql.DB) *Repository {
	pool := db

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) GetTimesheets(ctx context.Context, businessID int64, weekStart string) ([]domain.TimesheetEntry, error) {
	var (
		rows []timesheetQueryRow
		err  error
	)
	weekStartDate, nextWeekStart, rangeErr := timesheetWeekRange(weekStart)
	if errors.Is(rangeErr, errWeekStartNotMonday) {
		// The previous equality-to-derived-week predicate could never match a
		// non-Monday date. Preserve that empty-result behavior.
		return []domain.TimesheetEntry{}, nil
	}
	if rangeErr == nil && weekStart != "" {
		weekRows, queryErr := r.queries.GetTimesheetsForWeek(ctx, sqlc.GetTimesheetsForWeekParams{
			BusinessID: businessID,
			Date:       sql.NullTime{Time: weekStartDate, Valid: true},
			Date_2:     sql.NullTime{Time: nextWeekStart, Valid: true},
		})
		err = queryErr
		rows = make([]timesheetQueryRow, 0, len(weekRows))
		for _, row := range weekRows {
			rows = append(rows, timesheetQueryRow{row.StaffID, row.StaffName, row.WeekStart, row.TotalHours, row.RegularHours, row.OvertimeHours})
		}
	} else {
		allRows, queryErr := r.queries.GetTimesheets(ctx, businessID)
		err = queryErr
		rows = make([]timesheetQueryRow, 0, len(allRows))
		for _, row := range allRows {
			rows = append(rows, timesheetQueryRow{row.StaffID, row.StaffName, row.WeekStart, row.TotalHours, row.RegularHours, row.OvertimeHours})
		}
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get timesheets", logger.Error(err))
		return nil, err
	}

	entries := make([]domain.TimesheetEntry, 0, len(rows))
	for _, row := range rows {
		staffID := int64(0)
		if row.staffID.Valid {
			staffID = row.staffID.Int64
		}
		entries = append(entries, domain.TimesheetEntry{
			StaffID:       staffID,
			StaffName:     row.staffName,
			WeekStart:     formatWeekStart(row.weekStart),
			TotalHours:    sqlutil.DecimalToFloat(row.totalHours),
			RegularHours:  sqlutil.DecimalToFloat(row.regularHours),
			OvertimeHours: sqlutil.DecimalToFloat(row.overtimeHours),
		})
	}
	return entries, nil
}

func timesheetWeekRange(value string) (time.Time, time.Time, error) {
	if value == "" {
		return time.Time{}, time.Time{}, nil
	}
	weekStart, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if weekStart.Weekday() != time.Monday {
		return time.Time{}, time.Time{}, errWeekStartNotMonday
	}
	return weekStart, weekStart.AddDate(0, 0, 7), nil
}

func formatWeekStart(t time.Time) string {
	return t.Format("2006-01-02")
}
