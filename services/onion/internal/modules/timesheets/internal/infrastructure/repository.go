package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

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
	if weekStart == "" {
		rows, err := r.queries.GetTimesheets(ctx, businessID)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to get timesheets", logger.Error(err))
			return nil, err
		}
		entries := make([]domain.TimesheetEntry, 0, len(rows))
		for _, row := range rows {
			entries = append(entries, newTimesheetEntry(row.StaffID, row.StaffName, row.WeekStart, row.TotalHours, row.RegularHours, row.OvertimeHours))
		}
		return entries, nil
	}

	start, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		return r.GetTimesheets(ctx, businessID, "")
	}
	rows, err := r.queries.GetTimesheetsForWeek(ctx, sqlc.GetTimesheetsForWeekParams{
		BusinessID: businessID,
		Date:       sql.NullTime{Time: start, Valid: true},
		Date_2:     sql.NullTime{Time: start.AddDate(0, 0, 7), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get timesheets for week", logger.Error(err))
		return nil, err
	}
	entries := make([]domain.TimesheetEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, newTimesheetEntry(row.StaffID, row.StaffName, row.WeekStart, row.TotalHours, row.RegularHours, row.OvertimeHours))
	}
	return entries, nil
}

func newTimesheetEntry(staffID sql.NullInt64, staffName string, weekStart time.Time, totalHours, regularHours, overtimeHours any) domain.TimesheetEntry {
	id := int64(0)
	if staffID.Valid {
		id = staffID.Int64
	}
	return domain.TimesheetEntry{
		StaffID:       id,
		StaffName:     staffName,
		WeekStart:     formatWeekStart(weekStart),
		TotalHours:    parseDecimal(totalHours),
		RegularHours:  parseDecimal(regularHours),
		OvertimeHours: parseDecimal(overtimeHours),
	}
}

func formatWeekStart(t time.Time) string {
	return t.Format("2006-01-02")
}

func parseDecimal(v any) float64 {
	switch val := v.(type) {
	case nil:
		return 0
	case float64:
		return val
	case []byte:
		return parseDecimalString(string(val))
	case string:
		return parseDecimalString(val)
	default:
		return parseDecimalString(fmt.Sprintf("%v", val))
	}
}

func parseDecimalString(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse timesheet decimal from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}
