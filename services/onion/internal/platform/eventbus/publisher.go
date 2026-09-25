package eventbus

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	sqlc "github.com/voxire/onion/internal/platform/eventbus/internal/infrastructure/persistence/sqlc"
)

// Event types. These strings are the contract between a producer and whoever
// subscribes, so they are values rather than an enum: a registration stores one
// in hub_webhooks.events and an automation stores one in
// crm_automations.trigger_type, both as text in the database.
//
// TypeOrderCreated must stay equal to the automations module's
// TriggerOrderCreated. They cannot be a shared constant because that trigger
// lives under internal/modules/automations/internal, which the Go toolchain
// will not let a platform package import. If they ever drift the dispatcher
// silently matches nothing, so the dispatcher test seeds a crm_automations row
// using this constant rather than a literal.
const (
	TypeOrderCreated = "order_created"
)

// Subject types name what an event is about.
const (
	SubjectOrder = "order"
)

// Event is one thing that happened, recorded for anyone who wants to know.
type Event struct {
	// BusinessID is the tenant the event belongs to. Required: an event with
	// no tenant cannot be routed to anyone safely.
	BusinessID int64
	// Type is one of the type constants above.
	Type string
	// SubjectType and SubjectID identify the row the event is about.
	SubjectType string
	SubjectID   int64
	// Payload is marshalled to JSON. Keep it to what a subscriber needs in
	// order to act without reading the database back, and out of it anything
	// that would be stale by delivery time.
	Payload any
	// DedupeKey makes a retried producer idempotent. Left empty it defaults to
	// type:subject_type:subject_id, which is right whenever one subject
	// produces a given event once.
	DedupeKey string
}

// Publish records an event inside the caller's transaction.
//
// Taking a *sql.Tx rather than a *sql.DB is the whole point of the outbox: the
// event and the domain write it describes commit or roll back together. A
// producer that published after committing would lose events on a crash between
// the two, and one that published before would announce changes that never
// happened.
//
// Publishing does not deliver anything. The dispatcher expands the event into
// per-target queues on its next pass.
func Publish(ctx context.Context, tx *sql.Tx, event Event) error {
	if tx == nil {
		return fmt.Errorf("eventbus: publish %q: no transaction", event.Type)
	}
	if event.BusinessID <= 0 {
		return fmt.Errorf("eventbus: publish %q: business id is required", event.Type)
	}
	if event.Type == "" {
		return fmt.Errorf("eventbus: publish: event type is required")
	}
	if event.SubjectType == "" || event.SubjectID <= 0 {
		return fmt.Errorf("eventbus: publish %q: subject is required", event.Type)
	}

	payload := []byte("{}")
	if event.Payload != nil {
		encoded, err := json.Marshal(event.Payload)
		if err != nil {
			return fmt.Errorf("eventbus: publish %q: encoding payload: %w", event.Type, err)
		}
		payload = encoded
	}

	dedupeKey := event.DedupeKey
	if dedupeKey == "" {
		dedupeKey = fmt.Sprintf("%s:%s:%d", event.Type, event.SubjectType, event.SubjectID)
	}

	return sqlc.New(tx).InsertEvent(ctx, sqlc.InsertEventParams{
		BusinessID:  event.BusinessID,
		EventType:   event.Type,
		SubjectType: event.SubjectType,
		SubjectID:   event.SubjectID,
		Payload:     payload,
		DedupeKey:   dedupeKey,
	})
}
