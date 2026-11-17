// Package outbox records Flow audit facts for the platform event relay.
package outbox

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/outbox/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/eventbus"
)

// Recorder appends an audit fact after the Flow change has committed. Flow
// writes do not share one transaction with the fact, so the caller logs a
// failed append rather than undoing a change that already happened.
type Recorder struct{ queries *sqlc.Queries }

func NewRecorder(db *sql.DB) *Recorder { return &Recorder{queries: sqlc.New(db)} }

func (r *Recorder) RecordAudit(ctx context.Context, event audit.Event) error {
	actorID, ip := event.ActorID, ""
	if user, ok := auth.GetAuthUser(ctx); ok && user != nil {
		if user.UserID > 0 {
			actorID = user.UserID
		}
		ip = user.IP
	}
	record, err := eventbus.Encode(eventbus.Event{
		BusinessID: event.BusinessID, Type: audit.MutationRecorded,
		SubjectType: event.EntityType, SubjectID: event.EntityID,
		DedupeKey: "mutation:" + uuid.NewString(),
		Payload:   map[string]any{"app": "flow", "actor_id": actorID, "action": event.Action, "description": event.Description, "old_values": event.OldValues, "new_values": event.NewValues, "ip_address": ip},
	})
	if err != nil {
		return err
	}
	if len(record.DedupeKey) > 186 {
		return errors.New("flow fact dedupe key exceeds archive capacity")
	}
	return r.queries.InsertFact(ctx, sqlc.InsertFactParams{BusinessID: record.BusinessID, EventType: record.Type, SubjectType: record.SubjectType, SubjectID: record.SubjectID, Payload: record.Payload, DedupeKey: record.DedupeKey, OccurredAt: record.OccurredAt})
}

type Source struct{ queries *sqlc.Queries }

func NewSource(db *sql.DB) *Source { return &Source{queries: sqlc.New(db)} }

func (s *Source) Pending(ctx context.Context, limit int32) ([]eventbus.Event, error) {
	rows, err := s.queries.PendingFacts(ctx, limit)
	if err != nil {
		return nil, err
	}
	facts := make([]eventbus.Event, 0, len(rows))
	for _, row := range rows {
		facts = append(facts, eventbus.Event{ID: row.ID, BusinessID: row.BusinessID, Type: row.EventType, SubjectType: row.SubjectType, SubjectID: row.SubjectID, Payload: row.Payload, DedupeKey: "flow:" + row.DedupeKey, OccurredAt: row.OccurredAt})
	}
	return facts, nil
}

func (s *Source) MarkRelayed(ctx context.Context, id uint64) error {
	return s.queries.MarkRelayed(ctx, id)
}
