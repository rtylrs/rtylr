package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/categories/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()

	return &Repository{
		pool: pool,
	}
}

func (r *Repository) ListCategories(ctx context.Context, ownerID int64) ([]domain.Category, error) {
	query := `
		SELECT id, owner_id, name, description, created_at, updated_at
		FROM erp_categories
		WHERE owner_id = ?
		ORDER BY name ASC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list categories", logger.Error(err))

		return nil, err
	}

	defer rows.Close()

	var categories []domain.Category

	for rows.Next() {
		var c domain.Category
		if err := c.FromRow(rows); err != nil {
			logger.Ctx(ctx).Error("Failed to scan category row", logger.Error(err))

			return nil, err
		}

		categories = append(categories, c)
	}

	return categories, rows.Err()
}

func (r *Repository) CreateCategory(ctx context.Context, ownerID int64, input domain.CategoryInput) (int64, error) {
	query := `
		INSERT INTO erp_categories (owner_id, name, description)
		VALUES (?, ?, ?)`

	res, err := r.pool.ExecContext(ctx, query,
		ownerID,
		input.Name,
		nullString(input.Description),
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create category", logger.Error(err))

		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdateCategory(ctx context.Context, ownerID, id int64, input domain.CategoryInput) error {
	query := `
		UPDATE erp_categories
		SET name = ?, description = ?
		WHERE owner_id = ? AND id = ?`

	_, err := r.pool.ExecContext(ctx, query,
		input.Name,
		nullString(input.Description),
		ownerID,
		id,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update category", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) DeleteCategory(ctx context.Context, ownerID, id int64) error {
	query := `DELETE FROM erp_categories WHERE owner_id = ? AND id = ?`

	_, err := r.pool.ExecContext(ctx, query, ownerID, id)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete category", logger.Error(err))

		return err
	}

	return nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
