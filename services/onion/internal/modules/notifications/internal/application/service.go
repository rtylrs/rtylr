package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/notifications/internal/domain"
	"github.com/voxire/onion/internal/modules/notifications/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

// Notifications are per-user and hot-read; piggyback on UserPrefix so
// /invalidate-cache?type=user covers them too.
const notificationsTTL = 30 * time.Second

type repository interface {
	List(ctx context.Context, businessID, userID int64) ([]domain.Notification, error)
	MarkRead(ctx context.Context, userID, notificationID int64) error
	MarkAllRead(ctx context.Context, userID int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func notificationsKey(businessID, userID int64) string {
	return cacheKeys.Scoped("notifications", businessID, fmt.Sprintf("user:%d", userID))
}

func invalidateNotificationsScope(ctx context.Context, businessID, userID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.UserPrefix(userID))
		cache.DeleteByPattern(ctx, notificationsKey(businessID, userID))
	}
}

func (s *Service) List(ctx context.Context, businessID, userID int64) ([]domain.Notification, error) {
	return cacher.Of[[]domain.Notification]().GetOrComputeSingle(ctx,
		notificationsKey(businessID, userID),
		func(ctx context.Context) ([]domain.Notification, error) {
			return s.repo.List(ctx, businessID, userID)
		},
		cacher.WithExpiry(ctx, notificationsTTL),
	)
}

func (s *Service) MarkRead(ctx context.Context, businessID, userID, notificationID int64) error {
	if err := s.repo.MarkRead(ctx, userID, notificationID); err != nil {
		return err
	}
	invalidateNotificationsScope(ctx, businessID, userID)
	return nil
}

func (s *Service) MarkAllRead(ctx context.Context, businessID, userID int64) error {
	if err := s.repo.MarkAllRead(ctx, userID); err != nil {
		return err
	}
	invalidateNotificationsScope(ctx, businessID, userID)
	return nil
}
