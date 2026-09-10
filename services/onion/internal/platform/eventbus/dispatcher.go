// Package eventbus expands the platform event outbox into per-target work.
//
// One outbox, two consumers. A producer records that a domain action happened
// by writing a platform_events row inside its own transaction. This package
// reads events nobody has expanded yet, works out which webhook registrations
// and which automations want each one, and writes a queue row per interested
// target. It delivers nothing and runs nothing: sending a webhook and executing
// an automation are the consumers' jobs, and they read those queues.
//
// Splitting fan-out from delivery is what lets one tenant's dead endpoint fail
// on its own row rather than holding up everyone else's events.
package eventbus

import (
	"context"
	"database/sql"
	"fmt"

	sqlc "github.com/voxire/onion/internal/platform/eventbus/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// defaultBatchSize bounds one pass. Small enough that a pass is quick and a
// crash re-does little, large enough that a busy tenant is not starved. Each
// event is its own transaction regardless, so this does not govern lock hold
// time.
const defaultBatchSize = 100

// Dispatcher expands outbox events into the webhook and automation queues.
type Dispatcher struct {
	db        *sql.DB
	queries   *sqlc.Queries
	batchSize int32
}

// NewDispatcher builds a dispatcher over the shared pool.
func NewDispatcher(db *sql.DB) *Dispatcher {
	return &Dispatcher{db: db, queries: sqlc.New(db), batchSize: defaultBatchSize}
}

// DispatchBatch expands up to one batch of undispatched events and reports how
// many were expanded.
//
// An event that fails to expand is logged and skipped rather than aborting the
// batch, so one bad row cannot stall the queue for every tenant. It stays
// undispatched and is retried on the next pass, which is safe because expansion
// is idempotent.
func (d *Dispatcher) DispatchBatch(ctx context.Context) (int, error) {
	events, err := d.queries.ListUndispatchedEvents(ctx, d.batchSize)
	if err != nil {
		return 0, fmt.Errorf("listing undispatched events: %w", err)
	}

	dispatched := 0
	for _, event := range events {
		if err := d.dispatchOne(ctx, event); err != nil {
			logger.Ctx(ctx).Error("eventbus: failed to expand event",
				logger.Any("event_id", event.ID),
				logger.Any("business_id", event.BusinessID),
				logger.String("event_type", event.EventType),
				logger.Error(err))
			continue
		}
		dispatched++
	}
	return dispatched, nil
}

// dispatchOne expands a single event inside one transaction.
//
// The queue rows and the dispatched_at mark commit together. Without that an
// event could be marked dispatched having produced no work, which is
// unrecoverable because nothing would ever look at it again, or produce work
// and stay unmarked, which would fan out again on the next pass. The unique
// keys on both queues make the second case harmless, but the transaction means
// it does not arise.
func (d *Dispatcher) dispatchOne(ctx context.Context, event sqlc.ListUndispatchedEventsRow) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	q := d.queries.WithTx(tx)

	webhookIDs, err := q.ListWebhooksForEvent(ctx, sqlc.ListWebhooksForEventParams{
		BusinessID: event.BusinessID,
		EventType:  event.EventType,
	})
	if err != nil {
		return fmt.Errorf("listing webhooks: %w", err)
	}
	for _, webhookID := range webhookIDs {
		if err := q.EnqueueWebhookDelivery(ctx, sqlc.EnqueueWebhookDeliveryParams{
			BusinessID: event.BusinessID,
			WebhookID:  webhookID,
			EventID:    event.ID,
		}); err != nil {
			return fmt.Errorf("enqueueing webhook delivery for registration %d: %w", webhookID, err)
		}
	}

	automationIDs, err := q.ListAutomationsForEvent(ctx, sqlc.ListAutomationsForEventParams{
		BusinessID:  event.BusinessID,
		TriggerType: event.EventType,
	})
	if err != nil {
		return fmt.Errorf("listing automations: %w", err)
	}
	for _, automationID := range automationIDs {
		if err := q.EnqueueAutomationRun(ctx, sqlc.EnqueueAutomationRunParams{
			BusinessID:   event.BusinessID,
			AutomationID: automationID,
			EventID:      event.ID,
		}); err != nil {
			return fmt.Errorf("enqueueing automation run for automation %d: %w", automationID, err)
		}
	}

	// An event with no interested target is still dispatched. Leaving it
	// unmarked would make the dispatcher re-read it on every pass forever.
	if _, err := q.MarkEventDispatched(ctx, event.ID); err != nil {
		return fmt.Errorf("marking dispatched: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
