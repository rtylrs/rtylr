package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/booking_calendar/internal/domain"
	"github.com/voxire/onion/internal/modules/booking_calendar/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		pool:    db,
	}
}

func (r *Repository) FetchAppointments(ctx context.Context, businessID int64, from, to string) ([]domain.CalendarAppointment, error) {
	fromTime, err := parseCalendarBoundary(from, false)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse calendar from time", logger.Error(err))
		return nil, err
	}
	toTime, err := parseCalendarBoundary(to, true)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse calendar to time", logger.Error(err))
		return nil, err
	}

	rows, err := r.queries.FetchCalendarAppointments(ctx, sqlc.FetchCalendarAppointmentsParams{
		BusinessID:  businessID,
		StartTime:   fromTime,
		StartTime_2: toTime,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to fetch appointments", logger.Error(err))
		return nil, err
	}

	appts := make([]domain.CalendarAppointment, 0, len(rows))
	for _, row := range rows {
		appt := domain.CalendarAppointment{
			ID:         row.ID,
			BusinessID: row.BusinessID,
			StartTime:  row.StartTime.Format("2006-01-02T15:04:05Z07:00"),
			EndTime:    row.EndTime.Format("2006-01-02T15:04:05Z07:00"),
		}
		if row.ServiceID.Valid {
			appt.ServiceID = row.ServiceID.Int64
		}
		if row.CustomerName.Valid {
			appt.CustomerName = row.CustomerName.String
		}
		if row.Status.Valid {
			appt.Status = string(row.Status.BookingAppointmentsStatus)
		}
		appts = append(appts, appt)
	}

	return appts, nil
}

// parseCalendarBoundary accepts either an RFC3339 timestamp or a plain
// YYYY-MM-DD date. The tool executor and REST callers both send date-only
// values (from_date/to_date), which the old RFC3339-only parse rejected, so
// every calendar request failed.
//
// A date-only upper bound is widened to the start of the following day: the
// query window is half-open (start_time < to), and callers asking for
// to=2026-06-08 mean the whole of that day, not just its midnight.
func parseCalendarBoundary(value string, upperBound bool) (time.Time, error) {
	if ts, err := time.Parse(time.RFC3339, value); err == nil {
		return ts, nil
	}
	if day, err := time.Parse("2006-01-02", value); err == nil {
		if upperBound {
			return day.AddDate(0, 0, 1), nil
		}
		return day, nil
	}
	return time.Time{}, fmt.Errorf("calendar time %q must be RFC3339 or YYYY-MM-DD", value)
}
