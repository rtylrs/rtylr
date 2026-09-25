package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/booking_calendar/internal/domain"
	"github.com/voxire/onion/internal/modules/booking_calendar/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const bookingCalendarTTL = 1 * time.Minute

type Service struct {
	repo domain.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func (s *Service) GetCalendar(ctx context.Context, businessID int64, from, to string) ([]domain.CalendarDay, error) {
	key := cacheKeys.Scoped("booking_calendar", businessID, "cal|f="+from+"|t="+to)
	return cacher.Of[[]domain.CalendarDay]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.CalendarDay, error) {
			return s.fetchCalendar(ctx, businessID, from, to)
		},
		cacher.WithExpiry(ctx, bookingCalendarTTL),
	)
}

func (s *Service) fetchCalendar(ctx context.Context, businessID int64, from, to string) ([]domain.CalendarDay, error) {
	appts, err := s.repo.FetchAppointments(ctx, businessID, from, to)
	if err != nil {
		return nil, err
	}

	dayMap := make(map[string][]domain.CalendarAppointment, len(appts))
	for _, appt := range appts {
		dateKey := "unknown"
		if len(appt.StartTime) >= 10 {
			dateKey = appt.StartTime[:10]
		}
		dayMap[dateKey] = append(dayMap[dateKey], appt)
	}

	days := make([]domain.CalendarDay, 0, len(dayMap))
	for date, appts := range dayMap {
		days = append(days, domain.CalendarDay{Date: date, Appointments: appts})
	}
	return days, nil
}
