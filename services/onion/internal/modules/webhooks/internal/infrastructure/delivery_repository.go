package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/webhooks/internal/domain"
	"github.com/voxire/onion/internal/modules/webhooks/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// ErrDeliveryNotClaimed means another worker took the delivery first. Not a
// failure: several workers reading one due list is the expected shape.
var ErrDeliveryNotClaimed = errors.New("webhook delivery already claimed")

// ListDueDeliveries returns deliveries ready to send, oldest due first.
func (r *Repository) ListDueDeliveries(ctx context.Context, limit int32) ([]domain.Delivery, error) {
	rows, err := r.queries.ListDueWebhookDeliveries(ctx, limit)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list due webhook deliveries", logger.Error(err))
		return nil, err
	}
	deliveries := make([]domain.Delivery, 0, len(rows))
	for _, row := range rows {
		deliveries = append(deliveries, domain.Delivery{
			ID:          row.ID,
			BusinessID:  row.BusinessID,
			WebhookID:   row.WebhookID,
			EventID:     row.EventID,
			Attempts:    int(row.Attempts),
			MaxAttempts: int(row.MaxAttempts),
		})
	}
	return deliveries, nil
}

// ClaimDelivery takes ownership. ErrDeliveryNotClaimed means another worker
// won the race and this one should skip.
func (r *Repository) ClaimDelivery(ctx context.Context, deliveryID uint64) error {
	result, err := r.queries.ClaimWebhookDelivery(ctx, deliveryID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to claim webhook delivery", logger.Error(err))
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrDeliveryNotClaimed
	}
	return nil
}

// MarkDelivered records a successful send and the status the endpoint gave.
func (r *Repository) MarkDelivered(ctx context.Context, deliveryID uint64, responseStatus int) error {
	return r.queries.MarkWebhookDelivered(ctx, sqlc.MarkWebhookDeliveredParams{
		ResponseStatus: nullResponseStatus(responseStatus),
		ID:             deliveryID,
	})
}

// MarkFailed schedules a retry, or dead-letters once attempts are spent. The
// decision is made in SQL against the row's own max_attempts so it cannot
// drift from what the table says.
func (r *Repository) MarkFailed(ctx context.Context, deliveryID uint64, retryAfterSeconds int64, responseStatus int, cause string) error {
	return r.queries.MarkWebhookDeliveryFailed(ctx, sqlc.MarkWebhookDeliveryFailedParams{
		RetryAfterSeconds: retryAfterSeconds,
		ResponseStatus:    nullResponseStatus(responseStatus),
		LastError:         sql.NullString{String: truncateCause(cause), Valid: cause != ""},
		ID:                deliveryID,
	})
}

// MarkDead ends a delivery that no number of retries could complete.
func (r *Repository) MarkDead(ctx context.Context, deliveryID uint64, cause string) error {
	return r.queries.MarkWebhookDeliveryDead(ctx, sqlc.MarkWebhookDeliveryDeadParams{
		LastError: sql.NullString{String: truncateCause(cause), Valid: cause != ""},
		ID:        deliveryID,
	})
}

// RecoverAbandoned requeues deliveries left in flight by a worker that died,
// and reports how many. Returns them to a retryable state, so a delivery that
// had already reached the subscriber may reach it twice: webhooks are
// at-least-once and every request carries the ids needed to dedupe.
func (r *Repository) RecoverAbandoned(ctx context.Context, staleAfterSeconds int64) (int64, error) {
	result, err := r.queries.RecoverAbandonedWebhookDeliveries(ctx, staleAfterSeconds)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to recover abandoned webhook deliveries", logger.Error(err))
		return 0, err
	}
	return result.RowsAffected()
}

// GetRegistrationForDelivery loads the subscriber, scoped by the delivery's
// business as well as its id.
func (r *Repository) GetRegistrationForDelivery(ctx context.Context, webhookID, businessID int64) (domain.Registration, error) {
	row, err := r.queries.GetWebhookForDelivery(ctx, sqlc.GetWebhookForDeliveryParams{
		ID:         webhookID,
		BusinessID: businessID,
	})
	if err != nil {
		return domain.Registration{}, err
	}
	return domain.Registration{
		ID:       row.ID,
		URL:      row.Url,
		Secret:   row.Secret,
		IsActive: row.IsActive,
	}, nil
}

// GetEventForDelivery loads the event being delivered.
func (r *Repository) GetEventForDelivery(ctx context.Context, eventID uint64, businessID int64) (domain.Event, error) {
	row, err := r.queries.GetEventForDelivery(ctx, sqlc.GetEventForDeliveryParams{
		ID:         eventID,
		BusinessID: businessID,
	})
	if err != nil {
		return domain.Event{}, err
	}
	return domain.Event{
		ID:          row.ID,
		Type:        row.EventType,
		SubjectType: row.SubjectType,
		SubjectID:   row.SubjectID,
		Payload:     row.Payload,
		OccurredAt:  row.OccurredAt,
	}, nil
}

// RecordTriggered writes hub_webhooks.last_triggered_at, which the column has
// carried since it was added and nothing has ever set.
func (r *Repository) RecordTriggered(ctx context.Context, webhookID, businessID int64) error {
	return r.queries.RecordWebhookTriggered(ctx, sqlc.RecordWebhookTriggeredParams{
		ID:         webhookID,
		BusinessID: businessID,
	})
}

// nullResponseStatus leaves the column NULL when there was no HTTP response at
// all, which is different from a response carrying status 0.
func nullResponseStatus(status int) sql.NullInt16 {
	if status <= 0 {
		return sql.NullInt16{}
	}
	return sql.NullInt16{Int16: int16(status), Valid: true} // #nosec G115 -- an HTTP status is three digits.
}

func truncateCause(cause string) string {
	const limit = 500
	if len(cause) <= limit {
		return cause
	}
	return cause[:limit] + "..."
}
