package outbox

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/crm/internal/outbox/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/eventbus"
)

// Publish joins only a crm transaction. Encoding failures roll back the domain change.
func Publish(ctx context.Context, tx *sql.Tx, event eventbus.Event) error {
	if tx == nil {
		return errors.New("crm outbox requires a transaction")
	}
	record, err := eventbus.Encode(event)
	if err != nil {
		return err
	}
	if len(record.DedupeKey) > 187 {
		return errors.New("crm fact dedupe key exceeds archive capacity")
	}
	return sqlc.New(tx).InsertFact(ctx, sqlc.InsertFactParams{BusinessID: record.BusinessID, EventType: record.Type, SubjectType: record.SubjectType, SubjectID: record.SubjectID, Payload: record.Payload, DedupeKey: record.DedupeKey, OccurredAt: record.OccurredAt})
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
		facts = append(facts, eventbus.Event{ID: row.ID, BusinessID: row.BusinessID, Type: row.EventType, SubjectType: row.SubjectType, SubjectID: row.SubjectID, Payload: row.Payload, DedupeKey: "crm:" + row.DedupeKey, OccurredAt: row.OccurredAt})
	}
	return facts, nil
}
func (s *Source) MarkRelayed(ctx context.Context, id uint64) error {
	return s.queries.MarkRelayed(ctx, id)
}
