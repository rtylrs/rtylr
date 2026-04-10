package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/pos_modifiers/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Modifier, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, name, price_delta, COALESCE(group_name,''), is_required, created_at, updated_at
		 FROM pos_modifiers WHERE owner_id = ? ORDER BY group_name, name`, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list modifiers", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Modifier, 0)
	for rows.Next() {
		var m domain.Modifier
		if err := rows.Scan(&m.ID, &m.OwnerID, &m.Name, &m.PriceDelta, &m.GroupName, &m.IsRequired, &m.CreatedAt, &m.UpdatedAt); err != nil {
			logger.Ctx(ctx).Error("Failed to scan modifier", logger.Error(err))
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.ModifierInput) (int64, error) {
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO pos_modifiers (owner_id, name, price_delta, group_name, is_required)
		 VALUES (?, ?, ?, ?, ?)`,
		ownerID, input.Name, input.PriceDelta, nullString(input.GroupName), input.IsRequired,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create modifier", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, ownerID, id int64, input domain.ModifierInput) error {
	result, err := r.pool.ExecContext(ctx,
		`UPDATE pos_modifiers SET name = ?, price_delta = ?, group_name = ?, is_required = ?
		 WHERE id = ? AND owner_id = ?`,
		input.Name, input.PriceDelta, nullString(input.GroupName), input.IsRequired, id, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update modifier", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrModifierNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, ownerID, id int64) error {
	result, err := r.pool.ExecContext(ctx,
		"DELETE FROM pos_modifiers WHERE id = ? AND owner_id = ?", id, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete modifier", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrModifierNotFound
	}
	return nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
