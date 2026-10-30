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
type Repository struct{ q *sqlc.Queries }

func NewRepository(db *sql.DB) *Repository { return &Repository{q: sqlc.New(db)} }

// Layout is the person's saved layout; empty when they never changed it.
func (r *Repository) Layout(ctx context.Context, userID int64) (layout.Layout, error) {
	raw, err := r.q.GetLayout(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return layout.Layout{}, nil
	}
	if err != nil {
		return layout.Layout{}, err
	}
	var out layout.Layout
	if err := json.Unmarshal(raw, &out); err != nil {
		return layout.Layout{}, err
	}
	return out, nil
}

// SaveLayout stores the layout, or forgets it once it is all default again.
func (r *Repository) SaveLayout(ctx context.Context, userID int64, l layout.Layout) error {
	if l.Empty() {
		return r.q.DeleteLayout(ctx, userID)
	}
	raw, err := json.Marshal(l)
	if err != nil {
		return err
	}
	return r.q.SaveLayout(ctx, sqlc.SaveLayoutParams{UserID: userID, Layout: raw})
}
