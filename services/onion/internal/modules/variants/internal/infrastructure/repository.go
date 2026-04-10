package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/variants/internal/domain"
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

func (r *Repository) ListVariants(ctx context.Context, ownerID, productID int64) ([]domain.Variant, error) {
	query := `
		SELECT id, owner_id, product_id, name, sku, barcode,
		       cost_price, sale_price, stock_quantity, attributes, is_active,
		       created_at, updated_at
		FROM erp_product_variants
		WHERE owner_id = ?`

	args := []interface{}{ownerID}

	if productID > 0 {
		query += ` AND product_id = ?`
		args = append(args, productID)
	}

	query += ` ORDER BY name ASC`

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list variants", logger.Error(err))

		return nil, err
	}

	defer rows.Close()

	var variants []domain.Variant

	for rows.Next() {
		var v domain.Variant
		if err := v.FromRow(rows); err != nil {
			logger.Ctx(ctx).Error("Failed to scan variant row", logger.Error(err))

			return nil, err
		}

		variants = append(variants, v)
	}

	return variants, rows.Err()
}

func (r *Repository) CreateVariant(ctx context.Context, ownerID int64, input domain.VariantInput) (int64, error) {
	query := `
		INSERT INTO erp_product_variants
		(owner_id, product_id, name, sku, barcode, cost_price, sale_price, stock_quantity, attributes, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.pool.ExecContext(ctx, query,
		ownerID,
		input.ProductID,
		input.Name,
		nullString(input.SKU),
		nullString(input.Barcode),
		input.CostPrice,
		input.SalePrice,
		input.StockQuantity,
		nullString(input.Attributes),
		input.IsActive,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create variant", logger.Error(err))

		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdateVariant(ctx context.Context, ownerID, id int64, input domain.VariantInput) error {
	query := `
		UPDATE erp_product_variants
		SET product_id = ?, name = ?, sku = ?, barcode = ?,
		    cost_price = ?, sale_price = ?, stock_quantity = ?, attributes = ?, is_active = ?
		WHERE owner_id = ? AND id = ?`

	_, err := r.pool.ExecContext(ctx, query,
		input.ProductID,
		input.Name,
		nullString(input.SKU),
		nullString(input.Barcode),
		input.CostPrice,
		input.SalePrice,
		input.StockQuantity,
		nullString(input.Attributes),
		input.IsActive,
		ownerID,
		id,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update variant", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) DeleteVariant(ctx context.Context, ownerID, id int64) error {
	query := `DELETE FROM erp_product_variants WHERE owner_id = ? AND id = ?`

	_, err := r.pool.ExecContext(ctx, query, ownerID, id)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete variant", logger.Error(err))

		return err
	}

	return nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
