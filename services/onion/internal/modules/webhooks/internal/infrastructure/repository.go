package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/voxire/onion/internal/modules/webhooks/internal/domain"
	"github.com/voxire/onion/internal/modules/webhooks/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/eventbus"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	db      *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		db:      db,
	}
}

func (r *Repository) DispatchEvent(ctx context.Context, event eventbus.Event) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	queries := sqlc.New(tx)
	webhookIDs, err := queries.ListWebhooksForEvent(ctx, sqlc.ListWebhooksForEventParams{
		BusinessID: event.BusinessID,
		EventType:  event.Type,
	})
	if err != nil {
		return err
	}
	for _, webhookID := range webhookIDs {
		if err := queries.EnqueueWebhookDelivery(ctx, sqlc.EnqueueWebhookDeliveryParams{
			BusinessID: event.BusinessID,
			WebhookID:  webhookID,
			EventID:    event.ID,
		}); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// WebhookExistsInBusiness reports whether the row exists within the business. The write queries
// in this module are :exec and cannot report a zero-row match, so writes
// pre-check existence to avoid silent success on wrong/foreign ids.
func (r *Repository) WebhookExistsInBusiness(ctx context.Context, businessID, id int64) bool {
	var exists bool
	if err := r.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM hub_webhooks WHERE id = ? AND business_id = ?)",
		id, businessID,
	).Scan(&exists); err != nil {
		return false
	}
	return exists
}

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.Webhook, error) {
	rows, err := r.queries.ListWebhooks(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list webhooks", logger.Error(err))
		return nil, err
	}

	webhooks := make([]domain.Webhook, 0, len(rows))
	for _, row := range rows {
		webhooks = append(webhooks, rowToWebhook(row))
	}

	return webhooks, nil
}

func (r *Repository) Create(ctx context.Context, ownerID, businessID int64, url, secret string, events []string) (int64, error) {
	eventsJSON, err := json.Marshal(events)
	if err != nil {
		return 0, err
	}

	res, err := r.queries.CreateWebhook(ctx, sqlc.CreateWebhookParams{
		OwnerID:    ownerID,
		BusinessID: businessID,
		Url:        url,
		Events:     eventsJSON,
		Secret:     secret,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create webhook", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) Delete(ctx context.Context, businessID, webhookID int64) error {
	if !r.WebhookExistsInBusiness(ctx, businessID, webhookID) {
		return domain.ErrWebhookNotFound
	}
	err := r.queries.DeleteWebhook(ctx, sqlc.DeleteWebhookParams{
		ID:         webhookID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete webhook", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) RotateSecret(ctx context.Context, businessID, webhookID int64, secret string) error {
	res, err := r.queries.RotateWebhookSecret(ctx, sqlc.RotateWebhookSecretParams{Secret: secret, ID: webhookID, BusinessID: businessID})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to rotate webhook secret", logger.Error(err))
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrWebhookNotFound
	}
	return nil
}

func (r *Repository) Get(ctx context.Context, businessID, webhookID int64) (domain.Webhook, error) {
	row, err := r.queries.GetWebhook(ctx, sqlc.GetWebhookParams{
		ID:         webhookID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get webhook", logger.Error(err))
		return domain.Webhook{}, err
	}

	return rowToWebhook(row), nil
}

func (r *Repository) Update(ctx context.Context, businessID, webhookID int64, url string, events []string, isActive bool) error {
	if !r.WebhookExistsInBusiness(ctx, businessID, webhookID) {
		return domain.ErrWebhookNotFound
	}
	eventsJSON, err := json.Marshal(events)
	if err != nil {
		return err
	}

	err = r.queries.UpdateWebhook(ctx, sqlc.UpdateWebhookParams{
		Url:        url,
		Events:     eventsJSON,
		IsActive:   isActive,
		ID:         webhookID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update webhook", logger.Error(err))
		return err
	}

	return nil
}

func rowToWebhook(row sqlc.HubWebhook) domain.Webhook {
	var w domain.Webhook
	w.ID = row.ID
	w.OwnerID = row.OwnerID
	w.BusinessID = row.BusinessID
	w.URL = row.Url
	w.Secret = row.Secret
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
	return w
}
