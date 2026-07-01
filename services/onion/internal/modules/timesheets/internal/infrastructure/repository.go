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
	params := sqlc.GetTimesheetsParams{BusinessID: businessID}
	if weekStart != "" {
		if t, err := time.Parse("2006-01-02", weekStart); err == nil {
			params.WeekStart = sql.NullTime{Time: t, Valid: true}
		}
	}

	rows, err := r.queries.GetTimesheets(ctx, params)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get timesheets", logger.Error(err))
		return nil, err
	}

	entries := make([]domain.TimesheetEntry, 0, len(rows))
	for _, row := range rows {
		staffID := int64(0)
		if row.StaffID.Valid {
			staffID = row.StaffID.Int64
		}
		entries = append(entries, domain.TimesheetEntry{
			StaffID:       staffID,
			StaffName:     row.StaffName,
			WeekStart:     formatWeekStart(row.WeekStart),
			TotalHours:    parseDecimal(row.TotalHours),
			RegularHours:  parseDecimal(row.RegularHours),
			OvertimeHours: parseDecimal(row.OvertimeHours),
		})
	}
	return entries, nil
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
