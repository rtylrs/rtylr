package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/voxire/onion/internal/modules/webhooks/internal/domain"
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

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Webhook, error) {
	query := `
		SELECT id, owner_id, url, events, is_active, last_triggered_at, created_at, updated_at
		FROM hub_webhooks
		WHERE owner_id = ?
		ORDER BY created_at DESC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list webhooks", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var webhooks []domain.Webhook
	for rows.Next() {
		var w domain.Webhook
		var eventsJSON []byte
		if err := rows.Scan(&w.ID, &w.OwnerID, &w.URL, &eventsJSON, &w.IsActive, &w.LastTriggeredAt, &w.CreatedAt, &w.UpdatedAt); err != nil {
			logger.Ctx(ctx).Error("Failed to scan webhook row", logger.Error(err))
			return nil, err
		}
		if err := json.Unmarshal(eventsJSON, &w.Events); err != nil {
			w.Events = []string{}
		}
		webhooks = append(webhooks, w)
	}

	return webhooks, rows.Err()
}

func (r *Repository) Create(ctx context.Context, ownerID int64, url, secret string, events []string) (int64, error) {
	eventsJSON, err := json.Marshal(events)
	if err != nil {
		return 0, err
	}

	query := `
		INSERT INTO hub_webhooks (owner_id, url, events, secret)
		VALUES (?, ?, ?, ?)`

	res, err := r.pool.ExecContext(ctx, query, ownerID, url, eventsJSON, secret)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create webhook", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) Delete(ctx context.Context, ownerID, webhookID int64) error {
	query := `DELETE FROM hub_webhooks WHERE id = ? AND owner_id = ?`

	_, err := r.pool.ExecContext(ctx, query, webhookID, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete webhook", logger.Error(err))
		return err
	}

	return nil
}
