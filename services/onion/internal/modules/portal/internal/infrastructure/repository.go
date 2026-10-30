package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/voxire/onion/internal/modules/portal/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/modules/portal/internal/layout"
)

// Repository keeps each person's Portal layout.
type Repository struct {
	db *sql.DB
	q  *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db, q: sqlc.New(db)} }

// Layout is the person's saved layout; empty when they never changed it.
func (r *Repository) Layout(ctx context.Context, userID int64) (layout.Layout, error) {
	raw, err := r.q.GetLayout(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return layout.Layout{}, nil
	}
	if err != nil {
		return layout.Layout{}, err
	}
	return decode(raw)
}

// UpdateLayout edits the person's layout with their row locked, so two
// changes at once (Caret runs tool calls in parallel) can't undo each other.
func (r *Repository) UpdateLayout(ctx context.Context, userID int64, edit func(layout.Layout) (layout.Layout, error)) (layout.Layout, error) {
	if err := r.q.EnsureLayout(ctx, userID); err != nil {
		return layout.Layout{}, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return layout.Layout{}, err
	}
	defer tx.Rollback()
	q := r.q.WithTx(tx)
	raw, err := q.LockLayout(ctx, userID)
	if err != nil {
		return layout.Layout{}, err
	}
	current, err := decode(raw)
	if err != nil {
		return layout.Layout{}, err
	}
	next, err := edit(current)
	if err != nil {
		return layout.Layout{}, err
	}
	encoded, err := json.Marshal(next)
	if err != nil {
		return layout.Layout{}, err
	}
	if err := q.SetLayout(ctx, sqlc.SetLayoutParams{Layout: encoded, UserID: userID}); err != nil {
		return layout.Layout{}, err
	}
	return next, tx.Commit()
}

func decode(raw json.RawMessage) (layout.Layout, error) {
	var out layout.Layout
	if err := json.Unmarshal(raw, &out); err != nil {
		return layout.Layout{}, err
	}
	return out, nil
}
