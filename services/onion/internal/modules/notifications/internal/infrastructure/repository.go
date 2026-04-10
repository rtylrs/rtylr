package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/modules/notifications/internal/domain"
	"github.com/voxire/onion/internal/modules/notifications/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		queries: sqlc.New(dbclient.Pool()),
	}
}

func (r *Repository) List(ctx context.Context, userID int64) ([]domain.Notification, error) {
	rows, err := r.queries.ListNotifications(ctx, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list notifications", logger.Error(err))
		return nil, err
	}

	var notifications []domain.Notification
	for _, row := range rows {
		var n domain.Notification
		n.ID = row.ID
		n.OwnerID = row.OwnerID
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

func (r *Repository) MarkRead(ctx context.Context, userID, notificationID int64) error {
	err := r.queries.MarkNotificationRead(ctx, sqlc.MarkNotificationReadParams{
		ID:     notificationID,
		UserID: userID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to mark notification as read", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) MarkAllRead(ctx context.Context, userID int64) error {
	err := r.queries.MarkAllNotificationsRead(ctx, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to mark all notifications as read", logger.Error(err))
		return err
	}

	return nil
}
