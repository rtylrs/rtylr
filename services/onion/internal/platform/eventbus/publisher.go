package eventbus

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/platform/eventbus/internal/infrastructure/persistence/sqlc"
)

const TypeOrderCreated = "order_created"
const SubjectOrder = "order"

// Event is a tenant-scoped fact. Payload contains immutable reference data.
type Event struct {
	ID          uint64
	BusinessID  int64
	Type        string
	SubjectType string
	SubjectID   int64
	OccurredAt  time.Time
	Payload     any
	DedupeKey   string
}

type Reader interface {
	GetEvent(context.Context, uint64, int64) (Event, error)
}

type EncodedEvent struct {
	BusinessID  int64
	Type        string
	SubjectType string
	SubjectID   int64
	OccurredAt  time.Time
	Payload     []byte
	DedupeKey   string
}

// Encode validates the envelope without accessing persistence. The default key
// permits one fact per subject and type; recurring facts require explicit keys.
func Encode(event Event) (EncodedEvent, error) {
	if event.BusinessID <= 0 || event.Type == "" || event.SubjectType == "" || event.SubjectID <= 0 {
		return EncodedEvent{}, fmt.Errorf("eventbus: incomplete fact %q", event.Type)
	}
	payload, err := PayloadJSON(event.Payload)
	if err != nil {
		return EncodedEvent{}, fmt.Errorf("eventbus: encode %q: %w", event.Type, err)
	}
	key := event.DedupeKey
	if key == "" {
		key = fmt.Sprintf("%s:%s:%d", event.Type, event.SubjectType, event.SubjectID)
	}
	if len(key) > 191 {
		return EncodedEvent{}, fmt.Errorf("eventbus: dedupe key exceeds 191 bytes")
	}
	occurred := event.OccurredAt
	if occurred.IsZero() {
		occurred = time.Now().UTC()
	}
	return EncodedEvent{BusinessID: event.BusinessID, Type: event.Type, SubjectType: event.SubjectType, SubjectID: event.SubjectID, Payload: payload, DedupeKey: key, OccurredAt: occurred}, nil
}

// Publisher owns the archive commit. Producer transactions never enter it.
type Publisher struct{ queries *sqlc.Queries }

func NewPublisher(db *sql.DB) *Publisher { return &Publisher{queries: sqlc.New(db)} }
func (p *Publisher) Record(ctx context.Context, event Event) error {
	record, err := Encode(event)
	if err != nil {
		return err
	}
	return p.queries.InsertEvent(ctx, sqlc.InsertEventParams{BusinessID: record.BusinessID, EventType: record.Type, SubjectType: record.SubjectType, SubjectID: record.SubjectID, Payload: record.Payload, DedupeKey: record.DedupeKey, OccurredAt: record.OccurredAt})
}

// PayloadJSON preserves encoded JSON from persistence and encodes typed producer facts.
func PayloadJSON(value any) ([]byte, error) {
	var raw []byte
	switch value := value.(type) {
	case nil:
		raw = []byte("{}")
	case []byte:
		raw = value
	case json.RawMessage:
		raw = value
	default:
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		raw = encoded
	}
	if !json.Valid(raw) {
		return nil, fmt.Errorf("invalid fact JSON")
	}
	return raw, nil
}
func DecodePayload(value any, target any) error {
	raw, err := PayloadJSON(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}
