package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/one_on_ones/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.OneOnOne, error) {
	query := `
		SELECT id, owner_id, manager_id, staff_id, scheduled_at, notes, action_items, status, created_at
		FROM hr_one_on_ones
		WHERE owner_id = ?
		ORDER BY scheduled_at DESC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list one-on-ones", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var items []domain.OneOnOne
	for rows.Next() {
		var item domain.OneOnOne
		if err := item.FromRow(rows); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.OneOnOneInput) (int64, error) {
	query := `
		INSERT INTO hr_one_on_ones (owner_id, manager_id, staff_id, scheduled_at, notes, action_items, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	status := input.Status
	if status == "" {
		status = "scheduled"
	}

	res, err := r.pool.ExecContext(ctx, query,
		ownerID, input.ManagerID, input.StaffID, input.ScheduledAt,
		sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		sql.NullString{String: input.ActionItems, Valid: input.ActionItems != ""},
		status,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create one-on-one", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}
