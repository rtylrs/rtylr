package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/pos_combos/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Combo, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, name, COALESCE(description,''), price,
		        COALESCE(items,'[]'), is_active, created_at, updated_at
		 FROM pos_combos WHERE owner_id = ? ORDER BY name`, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list combos", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Combo, 0)
	for rows.Next() {
		var c domain.Combo
		if err := rows.Scan(&c.ID, &c.OwnerID, &c.Name, &c.Description, &c.Price, &c.Items, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			logger.Ctx(ctx).Error("Failed to scan combo", logger.Error(err))
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.ComboInput) (int64, error) {
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO pos_combos (owner_id, name, description, price, items, is_active)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		ownerID, input.Name, nullString(input.Description), input.Price, nullString(input.Items), input.IsActive,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create combo", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, ownerID, id int64, input domain.ComboInput) error {
	result, err := r.pool.ExecContext(ctx,
		`UPDATE pos_combos SET name = ?, description = ?, price = ?, items = ?, is_active = ?
		 WHERE id = ? AND owner_id = ?`,
		input.Name, nullString(input.Description), input.Price, nullString(input.Items), input.IsActive, id, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update combo", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrComboNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, ownerID, id int64) error {
	result, err := r.pool.ExecContext(ctx,
		"DELETE FROM pos_combos WHERE id = ? AND owner_id = ?", id, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete combo", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrComboNotFound
	}
	return nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
