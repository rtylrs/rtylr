package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/voxire/onion/internal/modules/webhooks/internal/domain"
	"github.com/voxire/onion/internal/modules/webhooks/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Webhook, error) {
	rows, err := r.queries.ListWebhooks(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list webhooks", logger.Error(err))
		return nil, err
	}

	var webhooks []domain.Webhook
	for _, row := range rows {
		var w domain.Webhook
		w.ID = row.ID
		w.OwnerID = row.OwnerID
		w.URL = row.Url
		w.IsActive = row.IsActive
		if row.LastTriggeredAt.Valid {
			w.LastTriggeredAt = &row.LastTriggeredAt.Time
		}
		if row.CreatedAt.Valid {
			w.CreatedAt = row.CreatedAt.Time
		}
		if row.UpdatedAt.Valid {
			w.UpdatedAt = row.UpdatedAt.Time
		}
		if err := json.Unmarshal(row.Events, &w.Events); err != nil {
			w.Events = []string{}
		}
		webhooks = append(webhooks, w)
	}

	return webhooks, nil
}

func (r *Repository) Create(ctx context.Context, ownerID int64, url, secret string, events []string) (int64, error) {
	eventsJSON, err := json.Marshal(events)
	if err != nil {
		return 0, err
	}

	res, err := r.queries.CreateWebhook(ctx, sqlc.CreateWebhookParams{
		OwnerID: ownerID,
		Url:     url,
		Events:  eventsJSON,
		Secret:  secret,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create webhook", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) Delete(ctx context.Context, ownerID, webhookID int64) error {
	err := r.queries.DeleteWebhook(ctx, sqlc.DeleteWebhookParams{
		ID:      webhookID,
		OwnerID: ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete webhook", logger.Error(err))
		return err
	}

	return nil
}
