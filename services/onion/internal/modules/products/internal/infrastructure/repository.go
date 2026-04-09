package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/voxire/onion/internal/modules/products/internal/domain"
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

func (r *Repository) ListProducts(ctx context.Context, ownerID int64, search string, categoryID, limit, offset int) ([]domain.Product, error) {
	query := `
		SELECT id, owner_id, category_id, supplier_id, name, sku, barcode, description,
		       unit, cost_price, sale_price, stock_quantity, low_stock_threshold,
		       reorder_quantity, is_active, track_inventory, created_at, updated_at
		FROM erp_products
		WHERE owner_id = ?`

	args := []interface{}{ownerID}

	if search != "" {
		query += ` AND (name LIKE ? OR sku LIKE ? OR barcode LIKE ?)`
		like := "%" + search + "%"
		args = append(args, like, like, like)
	}

	if categoryID > 0 {
		query += ` AND category_id = ?`
		args = append(args, categoryID)
	}

	query += ` ORDER BY name ASC`

	if limit > 0 {
		query += fmt.Sprintf(` LIMIT %d OFFSET %d`, limit, offset)
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list products", logger.Error(err))

		return nil, err
	}

	defer rows.Close()

	var products []domain.Product

	for rows.Next() {
		var p domain.Product
		if err := p.FromRow(rows); err != nil {
			logger.Ctx(ctx).Error("Failed to scan product row", logger.Error(err))

			return nil, err
		}

		products = append(products, p)
	}

	return products, rows.Err()
}

func (r *Repository) GetProduct(ctx context.Context, ownerID, id int64) (domain.Product, error) {
	query := `
		SELECT id, owner_id, category_id, supplier_id, name, sku, barcode, description,
		       unit, cost_price, sale_price, stock_quantity, low_stock_threshold,
		       reorder_quantity, is_active, track_inventory, created_at, updated_at
		FROM erp_products
		WHERE owner_id = ? AND id = ?`

	rows, err := r.pool.QueryContext(ctx, query, ownerID, id)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get product", logger.Error(err))

		return domain.Product{}, err
	}

	defer rows.Close()

	if !rows.Next() {
		return domain.Product{}, domain.ErrProductNotFound
	}

	var p domain.Product
	if err := p.FromRow(rows); err != nil {
		logger.Ctx(ctx).Error("Failed to scan product row", logger.Error(err))

		return domain.Product{}, err
	}

	return p, nil
}

func (r *Repository) CreateProduct(ctx context.Context, ownerID int64, input domain.ProductInput) (int64, error) {
	query := `
		INSERT INTO erp_products
		(owner_id, category_id, supplier_id, name, sku, barcode, description,
		 unit, cost_price, sale_price, low_stock_threshold, reorder_quantity,
		 is_active, track_inventory)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.pool.ExecContext(ctx, query,
		ownerID,
		nullInt64(input.CategoryID),
		nullInt64(input.SupplierID),
		input.Name,
		nullString(input.SKU),
		nullString(input.Barcode),
		nullString(input.Description),
		nullString(input.Unit),
		input.CostPrice,
		input.SalePrice,
		input.LowStockThreshold,
		input.ReorderQuantity,
		input.IsActive,
		input.TrackInventory,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create product", logger.Error(err))

		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdateProduct(ctx context.Context, ownerID, id int64, input domain.ProductInput) error {
	query := `
		UPDATE erp_products
		SET category_id = ?, supplier_id = ?, name = ?, sku = ?, barcode = ?,
		    description = ?, unit = ?, cost_price = ?, sale_price = ?,
		    low_stock_threshold = ?, reorder_quantity = ?, is_active = ?, track_inventory = ?
		WHERE owner_id = ? AND id = ?`

	_, err := r.pool.ExecContext(ctx, query,
		nullInt64(input.CategoryID),
		nullInt64(input.SupplierID),
		input.Name,
		nullString(input.SKU),
		nullString(input.Barcode),
		nullString(input.Description),
		nullString(input.Unit),
		input.CostPrice,
		input.SalePrice,
		input.LowStockThreshold,
		input.ReorderQuantity,
		input.IsActive,
		input.TrackInventory,
		ownerID,
		id,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update product", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) DeleteProduct(ctx context.Context, ownerID, id int64) error {
	query := `DELETE FROM erp_products WHERE owner_id = ? AND id = ?`

	_, err := r.pool.ExecContext(ctx, query, ownerID, id)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete product", logger.Error(err))

		return err
	}

	return nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}
