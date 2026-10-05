package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/activity/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/eventbus"
)

type mutationPayload struct {
	App         string          `json:"app"`
	ActorID     int64           `json:"actor_id"`
	Action      string          `json:"action"`
	Description string          `json:"description"`
	OldValues   json.RawMessage `json:"old_values"`
	NewValues   json.RawMessage `json:"new_values"`
	IPAddress   string          `json:"ip_address"`
}

// RecordFact commits the Activity projection independently. Event identity makes retries idempotent.
func (r *Repository) RecordFact(ctx context.Context, fact eventbus.Event) error {
	if fact.Type != audit.MutationRecorded {
		return nil
	}
	if fact.ID == 0 || fact.ID > math.MaxInt64 || fact.BusinessID <= 0 || fact.SubjectID <= 0 {
		return errors.New("invalid mutation fact identity")
	}
	var payload mutationPayload
	if err := eventbus.DecodePayload(fact.Payload, &payload); err != nil {
		return err
	}
	if (payload.App != "erp" && payload.App != "pos") || payload.Action == "" {
		return errors.New("invalid mutation fact payload")
	}
	occurred := fact.OccurredAt
	if occurred.IsZero() {
		occurred = time.Now().UTC()
	}
	return r.queries.RecordMutation(ctx, sqlc.RecordMutationParams{
		OwnerID: fact.BusinessID, BusinessID: fact.BusinessID, App: payload.App,
		EntityType: fact.SubjectType, EntityID: sql.NullInt64{Int64: fact.SubjectID, Valid: true},
		Action: payload.Action, Description: sql.NullString{String: payload.Description, Valid: payload.Description != ""},
		UserID:    sql.NullInt64{Int64: payload.ActorID, Valid: payload.ActorID > 0},
		OldValues: nullableJSON(payload.OldValues), NewValues: nullableJSON(payload.NewValues),
		IpAddress: sql.NullString{String: payload.IPAddress, Valid: payload.IPAddress != ""},
		CreatedAt: sql.NullTime{Time: occurred, Valid: true}, SourceEventID: sql.NullInt64{Int64: int64(fact.ID), Valid: true},
	})
}

func nullableJSON(value json.RawMessage) json.RawMessage {
	if string(value) == "null" {
		return nil
	}
	return value
}
