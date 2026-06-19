package infrastructure

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/booking_calendar/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		pool: db,
	}
}

func (r *Repository) FetchAppointments(ctx context.Context, businessID int64, from, to string) ([]domain.CalendarAppointment, error) {
	query, args, err := sq.Select("id", "service_id", "customer_name", "start_time", "end_time", "status", "business_id").
		From("booking_appointments").
		Where(sq.Eq{"business_id": businessID}).
		Where(sq.GtOrEq{"start_time": from}).
		Where(sq.Lt{"start_time": to}).
		OrderBy("start_time ASC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build fetch appointments query", logger.Error(err))
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to fetch appointments", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var appts []domain.CalendarAppointment
	for rows.Next() {
		var (
			id        int64
			serviceID sql.NullInt64
			custName  sql.NullString
			startTime time.Time
			endTime   time.Time
			status    sql.NullString
			bizID     int64
		)
		if err := rows.Scan(&id, &serviceID, &custName, &startTime, &endTime, &status, &bizID); err != nil {
			return nil, err
		}
		appt := domain.CalendarAppointment{
			ID:         id,
			BusinessID: bizID,
			StartTime:  startTime.Format("2006-01-02T15:04:05Z07:00"),
			EndTime:    endTime.Format("2006-01-02T15:04:05Z07:00"),
		}
		if serviceID.Valid {
			appt.ServiceID = serviceID.Int64
		}
		if custName.Valid {
			appt.CustomerName = custName.String
		}
		if status.Valid {
			appt.Status = status.String
		}
		appts = append(appts, appt)
	}

	return appts, rows.Err()
}
