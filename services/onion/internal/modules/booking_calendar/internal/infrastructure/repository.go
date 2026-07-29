package infrastructure

import (
	"context"
	"database/sql"
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
	fromTime, err := time.Parse(time.RFC3339, from)
	if err != nil {
		fromTime, err = time.Parse("2006-01-02T15:04:05Z07:00", from)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to parse calendar from time", logger.Error(err))
			return nil, err
		}
	}
	toTime, err := time.Parse(time.RFC3339, to)
	if err != nil {
		toTime, err = time.Parse("2006-01-02T15:04:05Z07:00", to)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to parse calendar to time", logger.Error(err))
			return nil, err
		}
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
