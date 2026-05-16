package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/booking_reminders/internal/domain"
	"github.com/voxire/onion/internal/modules/booking_reminders/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	bookingRemindersScope = "booking_reminders"
	bookingRemindersTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.Reminder, error)
	Create(ctx context.Context, ownerID int64, input domain.ReminderInput) (int64, error)
}

type Service struct{ repo repository }

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func invalidateBookingRemindersScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(bookingRemindersScope, ownerID))
	}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Reminder, error) {
	return cacher.Of[[]domain.Reminder]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(bookingRemindersScope, ownerID, "list"),
		func(ctx context.Context) ([]domain.Reminder, error) {
			return s.repo.List(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, bookingRemindersTTL),
	)
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.ReminderInput) (int64, error) {
	id, err := s.repo.Create(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}
	invalidateBookingRemindersScope(ctx, ownerID)
	return id, nil
}
