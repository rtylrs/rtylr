package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/voxire/onion/internal/modules/products/internal/domain"
	"github.com/voxire/onion/internal/modules/products/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) ListProducts(ctx context.Context, ownerID int64, search string, categoryID, limit, offset int) ([]domain.Product, error) {
	// ListProducts uses dynamic query building (search, categoryID, limit/offset)
	// which cannot be expressed as a single static sqlc query.
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
	row, err := r.queries.GetProduct(ctx, sqlc.GetProductParams{
		OwnerID: ownerID,
		ID:      id,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Product{}, domain.ErrProductNotFound
		}

		logger.Ctx(ctx).Error("Failed to get product", logger.Error(err))

		return domain.Product{}, err
	}

	return productFromRow(row), nil
}

// GetProductByBarcode powers the barcode-scan flow in POS + ERP. An
// empty / whitespace barcode is rejected at the service layer; at the
// repo level we trust the input and just match exactly. Owner-scoped
// so a barcode collision across tenants won't leak data.
func (r *Repository) GetProductByBarcode(ctx context.Context, ownerID int64, barcode string) (domain.Product, error) {
	// Dynamic query rather than sqlc: the SELECT shape matches
	// ListProducts exactly, so reusing productFromRow keeps parity
	// with GetProduct's output.
	const q = `
		SELECT id, owner_id, category_id, supplier_id, name, sku, barcode, description,
		       unit, cost_price, sale_price, stock_quantity, low_stock_threshold,
		       reorder_quantity, is_active, track_inventory, created_at, updated_at
		FROM erp_products
		WHERE owner_id = ? AND barcode = ?
		LIMIT 1`

	row := r.pool.QueryRowContext(ctx, q, ownerID, barcode)

	var p domain.Product
	var categoryID, supplierID sql.NullInt64
	var barcodeCol, description, unit sql.NullString
	var costPrice, salePrice, lowStockThreshold, reorderQuantity sql.NullFloat64

	err := row.Scan(
		&p.ID, &p.OwnerID, &categoryID, &supplierID, &p.Name, &p.SKU, &barcodeCol, &description,
		&unit, &costPrice, &salePrice, &p.StockQuantity, &lowStockThreshold,
		&reorderQuantity, &p.IsActive, &p.TrackInventory, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Product{}, domain.ErrProductNotFound
		}
		logger.Ctx(ctx).Error("Failed to look up product by barcode", logger.Error(err))
		return domain.Product{}, err
	}

	if categoryID.Valid {
		p.CategoryID = categoryID.Int64
	}
	if supplierID.Valid {
		p.SupplierID = supplierID.Int64
	}
	if barcodeCol.Valid {
		p.Barcode = barcodeCol.String
	}
	if description.Valid {
		p.Description = description.String
	}
	if unit.Valid {
		p.Unit = unit.String
	}
	if costPrice.Valid {
		p.CostPrice = costPrice.Float64
	}
	if salePrice.Valid {
		p.SalePrice = salePrice.Float64
	}
	if lowStockThreshold.Valid {
		p.LowStockThreshold = lowStockThreshold.Float64
	}
	if reorderQuantity.Valid {
		p.ReorderQuantity = reorderQuantity.Float64
	}
	return p, nil
}

func (r *Repository) CreateProduct(ctx context.Context, ownerID int64, input domain.ProductInput) (int64, error) {
	result, err := r.queries.CreateProduct(ctx, sqlc.CreateProductParams{
		OwnerID:           ownerID,
		CategoryID:        nullInt64(input.CategoryID),
		SupplierID:        nullInt64(input.SupplierID),
		Name:              input.Name,
		Sku:               nullString(input.SKU),
		Barcode:           nullString(input.Barcode),
		Description:       nullString(input.Description),
		Unit:              nullString(input.Unit),
		CostPrice:         strconv.FormatFloat(input.CostPrice, 'f', -1, 64),
		SalePrice:         strconv.FormatFloat(input.SalePrice, 'f', -1, 64),
		LowStockThreshold: nullStringFromFloat(input.LowStockThreshold),
		ReorderQuantity:   nullStringFromFloat(input.ReorderQuantity),
		IsActive:          sql.NullBool{Bool: input.IsActive, Valid: true},
		TrackInventory:    sql.NullBool{Bool: input.TrackInventory, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create product", logger.Error(err))

		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdateProduct(ctx context.Context, ownerID, id int64, input domain.ProductInput) error {
	err := r.queries.UpdateProduct(ctx, sqlc.UpdateProductParams{
		CategoryID:        nullInt64(input.CategoryID),
		SupplierID:        nullInt64(input.SupplierID),
		Name:              input.Name,
		Sku:               nullString(input.SKU),
		Barcode:           nullString(input.Barcode),
		Description:       nullString(input.Description),
		Unit:              nullString(input.Unit),
		CostPrice:         strconv.FormatFloat(input.CostPrice, 'f', -1, 64),
		SalePrice:         strconv.FormatFloat(input.SalePrice, 'f', -1, 64),
		LowStockThreshold: nullStringFromFloat(input.LowStockThreshold),
		ReorderQuantity:   nullStringFromFloat(input.ReorderQuantity),
		IsActive:          sql.NullBool{Bool: input.IsActive, Valid: true},
		TrackInventory:    sql.NullBool{Bool: input.TrackInventory, Valid: true},
		OwnerID:           ownerID,
		ID:                id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update product", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) DeleteProduct(ctx context.Context, ownerID, id int64) error {
	err := r.queries.DeleteProduct(ctx, sqlc.DeleteProductParams{
		OwnerID: ownerID,
		ID:      id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete product", logger.Error(err))

		return err
	}

	return nil
}

func productFromRow(row sqlc.ErpProduct) domain.Product {
	var p domain.Product
	p.ID = row.ID
	p.OwnerID = row.OwnerID
	p.Name = row.Name

	if row.CategoryID.Valid {
		p.CategoryID = row.CategoryID.Int64
	}

	if row.SupplierID.Valid {
		p.SupplierID = row.SupplierID.Int64
	}

	if row.Sku.Valid {
		p.SKU = row.Sku.String
	}

	if row.Barcode.Valid {
		p.Barcode = row.Barcode.String
	}

	if row.Description.Valid {
		p.Description = row.Description.String
	}

	if row.Unit.Valid {
		p.Unit = row.Unit.String
	}

	p.CostPrice, _ = strconv.ParseFloat(row.CostPrice, 64)
	p.SalePrice, _ = strconv.ParseFloat(row.SalePrice, 64)
	p.StockQuantity, _ = strconv.ParseFloat(row.StockQuantity, 64)

	if row.LowStockThreshold.Valid {
		p.LowStockThreshold, _ = strconv.ParseFloat(row.LowStockThreshold.String, 64)
	}

	if row.ReorderQuantity.Valid {
		p.ReorderQuantity, _ = strconv.ParseFloat(row.ReorderQuantity.String, 64)
	}

	if row.IsActive.Valid {
		p.IsActive = row.IsActive.Bool
	}

	if row.TrackInventory.Valid {
		p.TrackInventory = row.TrackInventory.Bool
	}

	if row.CreatedAt.Valid {
		p.CreatedAt = row.CreatedAt.Time
	}

	if row.UpdatedAt.Valid {
		p.UpdatedAt = row.UpdatedAt.Time
	}

	return p
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}

func nullStringFromFloat(v float64) sql.NullString {
	s := strconv.FormatFloat(v, 'f', -1, 64)
	return sql.NullString{String: s, Valid: true}
}
