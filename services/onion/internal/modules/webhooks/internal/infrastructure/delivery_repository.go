package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

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
	cause = strings.ToValidUTF8(cause, "")
	if len(cause) <= limit {
		return cause
	}
	cut := limit
	for !utf8.ValidString(cause[:cut]) {
		cut--
	}
	return cause[:cut] + "..."
}

// ListDeliveriesForBusiness returns a page of delivery attempts for one
// business, most recent first, with the total matching the same filter.
//
// Scoped by business in both statements, so a caller only ever sees the
// deliveries of registrations they own.
func (r *Repository) ListDeliveriesForBusiness(ctx context.Context, businessID int64, filter domain.DeliveryFilter) ([]domain.DeliveryRecord, int64, error) {
	total, err := r.queries.CountWebhookDeliveriesForBusiness(ctx, sqlc.CountWebhookDeliveriesForBusinessParams{
		BusinessID: businessID,
		WebhookID:  filter.WebhookID,
		Status:     sqlc.WebhookDeliveriesStatus(filter.Status),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count webhook deliveries", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListWebhookDeliveriesForBusiness(ctx, sqlc.ListWebhookDeliveriesForBusinessParams{
		BusinessID: businessID,
		WebhookID:  filter.WebhookID,
		// An unrecognised status is not rejected here: it simply matches
		// nothing, which is the right answer for a caller that asked for a
		// status that does not exist.
		Status: sqlc.WebhookDeliveriesStatus(filter.Status),
		Limit:  filter.Limit,
		Offset: int32(filter.Offset), // #nosec G115 -- bounded by the pagination limit the handler applies.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list webhook deliveries", logger.Error(err))
		return nil, 0, err
	}

	out := make([]domain.DeliveryRecord, 0, len(rows))
	for _, row := range rows {
		record := domain.DeliveryRecord{
			ID:          row.ID,
			WebhookID:   row.WebhookID,
			EventID:     row.EventID,
			EventType:   row.EventType,
			Status:      string(row.Status),
			Attempts:    int(row.Attempts),
			MaxAttempts: int(row.MaxAttempts),
			CreatedAt:   row.CreatedAt,
		}
		if row.ResponseStatus.Valid {
			record.ResponseStatus = int(row.ResponseStatus.Int16)
		}
		if row.LastError.Valid {
			record.LastError = row.LastError.String
		}
		if row.NextAttemptAt.IsZero() {
			record.NextAttemptAt = nil
		} else {
			next := row.NextAttemptAt
			record.NextAttemptAt = &next
		}
		if row.DeliveredAt.Valid {
			delivered := row.DeliveredAt.Time
			record.DeliveredAt = &delivered
		}
		out = append(out, record)
	}
	return out, total, nil
}

// ListFailingRegistrations returns registrations that dead-lettered something
// within the window, one row each rather than one per lost event.
//
// notifiedOn excludes registrations already told on that UTC day, which is what
// lets successive batches advance past the head of the failing set.
func (r *Repository) ListFailingRegistrations(ctx context.Context, withinHours int64, limit int32, notifiedOn string) ([]domain.FailingRegistration, error) {
	rows, err := r.queries.ListRegistrationsWithRecentDeadDeliveries(ctx, sqlc.ListRegistrationsWithRecentDeadDeliveriesParams{
		WithinHours: withinHours,
		NotifiedOn:  notifiedOn,
		Limit:       limit,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list registrations with dead deliveries", logger.Error(err))
		return nil, err
	}

	out := make([]domain.FailingRegistration, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.FailingRegistration{
			BusinessID: row.BusinessID,
			WebhookID:  row.WebhookID,
			URL:        row.Url,
			DeadCount:  row.DeadCount,
			LastDeadAt: asTime(row.LastDeadAt),
		})
	}
	return out, nil
}

// GetNotificationRecipient resolves who to tell about a failing registration.
//
// sql.ErrNoRows means there is nobody to tell: the creator has left the
// business, or their membership is no longer active. The caller logs that
// rather than falling back to somebody who never asked for the message.
func (r *Repository) GetNotificationRecipient(ctx context.Context, webhookID, businessID int64) (domain.NotificationRecipient, error) {
	row, err := r.queries.GetWebhookNotificationRecipient(ctx, sqlc.GetWebhookNotificationRecipientParams{
		ID:         webhookID,
		BusinessID: businessID,
	})
	if err != nil {
		return domain.NotificationRecipient{}, err
	}
	return domain.NotificationRecipient{
		UserID: row.ID,
		Email:  row.Email,
		Name:   row.FullName,
	}, nil
}

// asTime narrows an aggregate column sqlc types as interface{}.
//
// MAX(updated_at) has no single declared type, so the driver may hand back a
// time.Time or the raw bytes depending on parseTime. An unrecognised shape
// yields the zero time rather than failing the whole listing: the value only
// decorates a notification, and losing it is not worth dropping the alert
// that something is failing.
func asTime(value any) time.Time {
	switch typed := value.(type) {
	case time.Time:
		return typed
	case []byte:
		if parsed, err := time.Parse("2006-01-02 15:04:05", string(typed)); err == nil {
			return parsed
		}
	case string:
		if parsed, err := time.Parse("2006-01-02 15:04:05", typed); err == nil {
			return parsed
		}
	}
	return time.Time{}
}
