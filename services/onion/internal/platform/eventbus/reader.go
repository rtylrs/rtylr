package eventbus

import (
	"context"
	"database/sql"
	"errors"

	sqlc "github.com/voxire/onion/internal/platform/eventbus/internal/infrastructure/persistence/sqlc"
)

var ErrEventNotFound = errors.New("event not found")

type SQLReader struct{ queries *sqlc.Queries }

func NewSQLReader(db *sql.DB) *SQLReader {
	return &SQLReader{queries: sqlc.New(db)}
}

func (r *SQLReader) GetEvent(ctx context.Context, eventID uint64, businessID int64) (Event, error) {
	row, err := r.queries.GetEvent(ctx, sqlc.GetEventParams{ID: eventID, BusinessID: businessID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Event{}, ErrEventNotFound
		}
		return Event{}, err
	}
	return Event{
		ID:          row.ID,
		BusinessID:  row.BusinessID,
		Type:        row.EventType,
		SubjectType: row.SubjectType,
		SubjectID:   row.SubjectID,
		OccurredAt:  row.OccurredAt,
		Payload:     []byte(row.Payload),
	}, nil
}
