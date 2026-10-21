package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/booking/reminders/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	bookingRemindersScope = "booking_reminders"
	bookingRemindersTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.Reminder, error)
	Create(ctx context.Context, businessID int64, input domain.ReminderInput) (int64, error)
	AppointmentExists(ctx context.Context, businessID, appointmentID int64) (bool, error)
}

type Service struct{ repo repository }

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func invalidateBookingRemindersScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(bookingRemindersScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Reminder, error) {
	return cacher.Of[[]domain.Reminder]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(bookingRemindersScope, businessID, "list")+"|domain=v2",
		func(ctx context.Context) ([]domain.Reminder, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, bookingRemindersTTL),
	)
}

func (s *Service) Create(ctx context.Context, businessID int64, input domain.ReminderInput) (int64, error) {
	// Only create reminders for appointments that exist and belong to the
	// requesting business. This prevents probing other tenants' appointment IDs.
	exists, err := s.repo.AppointmentExists(ctx, businessID, input.AppointmentID)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, domain.ErrAppointmentNotFound
	}
	id, err := s.repo.Create(ctx, businessID, input)
	if err != nil {
		return 0, err
	}
	invalidateBookingRemindersScope(ctx, businessID)
	return id, nil
}
