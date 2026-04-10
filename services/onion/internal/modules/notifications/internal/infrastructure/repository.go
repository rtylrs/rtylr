package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/notifications/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		pool: dbclient.Pool(),
	}
}

func (r *Repository) List(ctx context.Context, userID int64) ([]domain.Notification, error) {
	query := `
		SELECT id, owner_id, user_id, title, COALESCE(body,''), type, is_read, COALESCE(link,''), created_at
		FROM hub_notifications
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT 100`

	rows, err := r.pool.QueryContext(ctx, query, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list notifications", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var notifications []domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.OwnerID, &n.UserID, &n.Title, &n.Body, &n.Type, &n.IsRead, &n.Link, &n.CreatedAt); err != nil {
			logger.Ctx(ctx).Error("Failed to scan notification row", logger.Error(err))
			return nil, err
		}
		notifications = append(notifications, n)
	}

	return notifications, rows.Err()
}

func (r *Repository) MarkRead(ctx context.Context, userID, notificationID int64) error {
	query := `UPDATE hub_notifications SET is_read = TRUE WHERE id = ? AND user_id = ?`

	_, err := r.pool.ExecContext(ctx, query, notificationID, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to mark notification as read", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) MarkAllRead(ctx context.Context, userID int64) error {
	query := `UPDATE hub_notifications SET is_read = TRUE WHERE user_id = ? AND is_read = FALSE`

	_, err := r.pool.ExecContext(ctx, query, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to mark all notifications as read", logger.Error(err))
		return err
	}

	return nil
}
