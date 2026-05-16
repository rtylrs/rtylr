package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/booking_calendar/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

const bookingCalendarTTL = 1 * time.Minute

type Service struct{ pool *sql.DB }

func NewService(_ context.Context) *Service {
	return &Service{pool: dbclient.Pool()}
}

func (s *Service) GetCalendar(ctx context.Context, ownerID int64, from, to string) ([]domain.CalendarDay, error) {
	key := cacheKeys.Scoped("booking_calendar", ownerID, fmt.Sprintf("cal|f=%s|t=%s", from, to))
	return cacher.Of[[]domain.CalendarDay]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.CalendarDay, error) {
			return s.fetchCalendar(ctx, ownerID, from, to)
		},
		cacher.WithExpiry(ctx, bookingCalendarTTL),
	)
}

func (s *Service) fetchCalendar(ctx context.Context, ownerID int64, from, to string) ([]domain.CalendarDay, error) {
	query := `SELECT id, service_id, customer_name, start_time, end_time, status
		FROM booking_appointments WHERE owner_id = ? AND start_time >= ? AND start_time < ? ORDER BY start_time ASC`

	rows, err := s.pool.QueryContext(ctx, query, ownerID, from, to)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get calendar", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	dayMap := make(map[string][]domain.CalendarAppointment)
	for rows.Next() {
		var (
			id        int64
			serviceID sql.NullInt64
			custName  sql.NullString
			startTime time.Time
			endTime   time.Time
			status    sql.NullString
		)
		if err := rows.Scan(&id, &serviceID, &custName, &startTime, &endTime, &status); err != nil {
			return nil, err
		}
		dateKey := startTime.Format("2006-01-02")
		appt := domain.CalendarAppointment{
			ID:        id,
			StartTime: startTime.Format("2006-01-02T15:04:05Z07:00"),
			EndTime:   endTime.Format("2006-01-02T15:04:05Z07:00"),
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
		dayMap[dateKey] = append(dayMap[dateKey], appt)
	}

	var days []domain.CalendarDay
	for date, appts := range dayMap {
		days = append(days, domain.CalendarDay{Date: date, Appointments: appts})
	}
	return days, rows.Err()
}
