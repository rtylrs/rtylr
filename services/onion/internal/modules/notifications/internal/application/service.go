package application

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/notifications/internal/domain"
	"github.com/voxire/onion/internal/modules/notifications/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/cursor"
)

// Notifications are per-user and hot-read; piggyback on UserPrefix so
// /invalidate-cache?type=user covers them too.
const notificationsTTL = 30 * time.Second

type repository interface {
	List(ctx context.Context, businessID, userID int64) ([]domain.Notification, error)
	MarkRead(ctx context.Context, userID, notificationID int64) error
	MarkAllRead(ctx context.Context, userID int64) error
}

type pagedRepository interface {
	ListPage(ctx context.Context, businessID, userID int64, limit int32, cursorCreatedAt *time.Time, cursorID *int64) ([]domain.Notification, error)
}

const (
	defaultPageSize int32 = 100
	maxPageSize     int32 = 100
)

type NotificationPage struct {
	Notifications []domain.Notification
	NextCursor    string
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func notificationsKey(businessID, userID int64, limit int32, cursorID *int64) string {
	cursor := ""
	if cursorID != nil {
		cursor = strconv.FormatInt(*cursorID, 10)
	}
	return cacheKeys.Scoped("notifications", businessID, fmt.Sprintf("user:%d|limit:%d|cursor:%s", userID, limit, cursor))
}

func invalidateNotificationsScope(ctx context.Context, businessID, userID int64) {
	cacher.InvalidateScope(ctx,
		cacheKeys.UserPrefix(userID),
		cacheKeys.Scoped("notifications", businessID, fmt.Sprintf("user:%d*", userID)),
	)
}

func (s *Service) List(ctx context.Context, businessID, userID int64) ([]domain.Notification, error) {
	page, err := s.ListPage(ctx, businessID, userID, 0, "")
	return page.Notifications, err
}

func (s *Service) ListPage(ctx context.Context, businessID, userID int64, limit int32, token string) (NotificationPage, error) {
	if limit <= 0 {
		limit = defaultPageSize
	}
	if limit > maxPageSize {
		limit = maxPageSize
	}

	var createdAt *time.Time
	var id *int64
	if token != "" {
		boundary, err := cursor.DecodeCreatedAtID(token)
		if err != nil {
			return NotificationPage{}, err
		}
		createdAt, id = &boundary.CreatedAt, &boundary.ID
	}

	return cacher.Of[NotificationPage]().GetOrComputeSingle(ctx,
		notificationsKey(businessID, userID, limit, id),
		func(ctx context.Context) (NotificationPage, error) {
			var rows []domain.Notification
			var err error
			if repo, ok := s.repo.(pagedRepository); ok {
				rows, err = repo.ListPage(ctx, businessID, userID, limit+1, createdAt, id)
			} else {
				rows, err = s.repo.List(ctx, businessID, userID)
			}
			if err != nil {
				return NotificationPage{}, err
			}
			page := NotificationPage{Notifications: rows}
			if int32(len(rows)) > limit { // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
				last := rows[limit-1]
				page.Notifications = rows[:limit]
				page.NextCursor = cursor.EncodeCreatedAtID(cursor.CreatedAtID{CreatedAt: last.CreatedAt, ID: last.ID})
			}
			return page, nil
		}, cacher.WithExpiry(ctx, notificationsTTL))
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
