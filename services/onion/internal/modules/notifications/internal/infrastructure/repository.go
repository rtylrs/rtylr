package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/notifications/internal/domain"
	"github.com/voxire/onion/internal/modules/notifications/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
	}
}

func (r *Repository) Create(ctx context.Context, notification domain.Notification) error {
	return r.queries.CreateNotification(ctx, sqlc.CreateNotificationParams{
		OwnerID:    notification.OwnerID,
		BusinessID: notification.BusinessID,
		UserID:     notification.UserID,
		Title:      notification.Title,
		Body:       sql.NullString{String: notification.Body, Valid: notification.Body != ""},
		Type:       notification.Type,
		Link:       sql.NullString{String: notification.Link, Valid: notification.Link != ""},
	})
}

func (r *Repository) List(ctx context.Context, businessID, userID int64) ([]domain.Notification, error) {
	return r.ListPage(ctx, businessID, userID, 100, nil, nil)
}

func (r *Repository) ListPage(ctx context.Context, businessID, userID int64, limit int32, cursorCreatedAt *time.Time, cursorID *int64) ([]domain.Notification, error) {
	rows, err := r.queries.ListNotifications(ctx, sqlc.ListNotificationsParams{
		BusinessID:      businessID,
		UserID:          userID,
		Limit:           limit,
		CursorCreatedAt: nullTime(cursorCreatedAt),
		CursorID:        nullInt64(cursorID),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list notifications", logger.Error(err))
		return nil, err
	}

	notifications := make([]domain.Notification, 0, len(rows))
	for _, row := range rows {
		var n domain.Notification
		n.ID = row.ID
		n.OwnerID = row.OwnerID
		n.BusinessID = row.BusinessID
		n.UserID = row.UserID
		n.Title = row.Title
		n.Body = row.Body
		n.Type = row.Type
		n.IsRead = row.IsRead
		n.Link = row.Link
		if row.CreatedAt.Valid {
			n.CreatedAt = row.CreatedAt.Time
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}

func nullTime(value *time.Time) sql.NullTime {
	if value == nil {
		return sql.NullTime{}
	}

	return sql.NullTime{Time: *value, Valid: true}
}

func nullInt64(value *int64) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}

	return sql.NullInt64{Int64: *value, Valid: true}
}

func (r *Repository) MarkRead(ctx context.Context, businessID, userID, notificationID int64) error {
	err := r.queries.MarkNotificationRead(ctx, sqlc.MarkNotificationReadParams{
		ID:         notificationID,
		BusinessID: businessID,
		UserID:     userID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to mark notification as read", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) MarkAllRead(ctx context.Context, businessID, userID int64) error {
	err := r.queries.MarkAllNotificationsRead(ctx, sqlc.MarkAllNotificationsReadParams{
		BusinessID: businessID,
		UserID:     userID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to mark all notifications as read", logger.Error(err))
		return err
	}

	return nil
}
