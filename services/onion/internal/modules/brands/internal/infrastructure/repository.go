package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/brands/internal/domain"
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

func (r *Repository) ListBrands(ctx context.Context, ownerID int64) ([]domain.Brand, error) {
	query := `
		SELECT b.id, b.owner_id, b.name, b.logo_url, b.description,
		       (SELECT COUNT(*) FROM erp_products p WHERE p.owner_id = b.owner_id AND p.brand_id = b.id) AS product_count,
		       b.created_at, b.updated_at
		FROM erp_brands b
		WHERE b.owner_id = ?
		ORDER BY b.name ASC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list brands", logger.Error(err))

		return nil, err
	}

	defer rows.Close()

	var brands []domain.Brand

	for rows.Next() {
		var b domain.Brand
		if err := b.FromRow(rows); err != nil {
			logger.Ctx(ctx).Error("Failed to scan brand row", logger.Error(err))

			return nil, err
		}

		brands = append(brands, b)
	}

	return brands, rows.Err()
}

func (r *Repository) CreateBrand(ctx context.Context, ownerID int64, input domain.BrandInput) (int64, error) {
	query := `
		INSERT INTO erp_brands (owner_id, name, logo_url, description)
		VALUES (?, ?, ?, ?)`

	res, err := r.pool.ExecContext(ctx, query,
		ownerID,
		input.Name,
		nullString(input.LogoURL),
		nullString(input.Description),
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create brand", logger.Error(err))

		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdateBrand(ctx context.Context, ownerID, id int64, input domain.BrandInput) error {
	query := `
		UPDATE erp_brands
		SET name = ?, logo_url = ?, description = ?
		WHERE owner_id = ? AND id = ?`

	_, err := r.pool.ExecContext(ctx, query,
		input.Name,
		nullString(input.LogoURL),
		nullString(input.Description),
		ownerID,
		id,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update brand", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) DeleteBrand(ctx context.Context, ownerID, id int64) error {
	query := `DELETE FROM erp_brands WHERE owner_id = ? AND id = ?`

	_, err := r.pool.ExecContext(ctx, query, ownerID, id)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete brand", logger.Error(err))

		return err
	}

	return nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
