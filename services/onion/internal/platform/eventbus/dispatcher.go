// Package eventbus publishes and distributes platform event facts.
package eventbus

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sqlc "github.com/voxire/onion/internal/platform/eventbus/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// defaultBatchSize bounds one pass. Small enough that a pass is quick and a
// crash re-does little, large enough that a busy tenant is not starved. Each
// event is its own transaction regardless, so this does not govern lock hold
// time.
const defaultBatchSize = 100

// Dispatcher publishes each undispatched fact to its registered consumers.
type Dispatcher struct {
	queries     *sqlc.Queries
	publisher   *Publisher
	sources     []FactSource
	batchSize   int32
	subscribers []EventSubscriber
}

type EventSubscriber interface {
	DispatchEvent(context.Context, Event) error
}

// NewDispatcher builds a dispatcher over the event log and its subscribers.
func NewDispatcher(db *sql.DB, subscribers ...EventSubscriber) *Dispatcher {
	return &Dispatcher{queries: sqlc.New(db), publisher: NewPublisher(db), batchSize: defaultBatchSize, subscribers: subscribers}
}

// DispatchBatch distributes up to one batch of undispatched events.
//
// An event that fails to distribute is logged and skipped rather than aborting the
// batch, so one bad row cannot stall the queue for every tenant. It stays
// undispatched and is retried on the next pass, which is safe because expansion
// is idempotent.
func (d *Dispatcher) DispatchBatch(ctx context.Context) (int, error) {
	relayErr := d.relay(ctx)
	events, err := d.queries.ListUndispatchedEvents(ctx, d.batchSize)
	if err != nil {
		return 0, fmt.Errorf("listing undispatched events: %w", err)
	}

	dispatched := 0
	for _, event := range events {
		if err := d.dispatchOne(ctx, event); err != nil {
			logger.Ctx(ctx).Error("eventbus: failed to distribute event",
				logger.Any("event_id", event.ID),
				logger.Any("business_id", event.BusinessID),
				logger.String("event_type", event.EventType),
				logger.Error(err))
			continue
		}
		dispatched++
	}
	return dispatched, relayErr
}

// dispatchOne asks each owning module to create local work before marking the
// fact complete. Consumer writes are idempotent, so partial fan-out can retry.
func (d *Dispatcher) dispatchOne(ctx context.Context, event sqlc.ListUndispatchedEventsRow) error {
	fact := Event{
		ID:          event.ID,
		BusinessID:  event.BusinessID,
		Type:        event.EventType,
		SubjectType: event.SubjectType,
		SubjectID:   event.SubjectID,
		Payload:     []byte(event.Payload),
		OccurredAt:  event.OccurredAt,
	}
	for _, subscriber := range d.subscribers {
		if err := subscriber.DispatchEvent(ctx, fact); err != nil {
			return fmt.Errorf("dispatching event to subscriber: %w", err)
		}
	}

	// Subscriber queues are independently committed and idempotent by event ID.
	// Mark only after each accepts the fact, so a partial fan-out can safely retry.
	if _, err := d.queries.MarkEventDispatched(ctx, event.ID); err != nil {
		return fmt.Errorf("marking dispatched: %w", err)
	}
	return nil
}

// FactSource grants committed reads and local acknowledgement, with no transaction capability.
type FactSource interface {
	Pending(context.Context, int32) ([]Event, error)
	MarkRelayed(context.Context, uint64) error
}

func (d *Dispatcher) WithSources(sources ...FactSource) *Dispatcher { d.sources = sources; return d }

func (d *Dispatcher) relay(ctx context.Context) error {
	var failures []error
	for _, source := range d.sources {
		facts, err := source.Pending(ctx, d.batchSize)
		if err != nil {
			failures = append(failures, fmt.Errorf("read producer facts: %w", err))
			continue
		}
		for _, fact := range facts {
			if err := d.publisher.Record(ctx, fact); err != nil {
				failures = append(failures, fmt.Errorf("archive producer fact: %w", err))
				break
			}
			if err := source.MarkRelayed(ctx, fact.ID); err != nil {
				failures = append(failures, fmt.Errorf("acknowledge producer fact: %w", err))
				break
			}
		}
	}
	return errors.Join(failures...)
}
