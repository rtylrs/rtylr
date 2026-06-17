package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/products/internal/domain"
	"github.com/voxire/onion/internal/modules/products/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/stock_core"
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

func (r *Repository) ListProducts(ctx context.Context, ownerID int64, search string, categoryID, limit, offset int) ([]domain.Product, bool, error) {
	qb := squirrel.Select(
		"id", "owner_id", "category_id", "supplier_id", "name", "sku", "barcode", "description",
		"unit", "cost_price", "sale_price", "stock_quantity", "low_stock_threshold",
		"reorder_quantity", "is_active", "track_inventory", "created_at", "updated_at",
	).
		From("erp_products").
		Where(squirrel.Eq{"owner_id": ownerID})

	if search != "" {
		like := "%" + search + "%"
		qb = qb.Where(squirrel.Or{
			squirrel.Like{"name": like},
			squirrel.Like{"sku": like},
			squirrel.Like{"barcode": like},
		})
	}

	if categoryID > 0 {
		qb = qb.Where(squirrel.Eq{"category_id": categoryID})
	}

	qb = qb.OrderBy("name ASC").Limit(uint64(limit + 1)).Offset(uint64(offset))

	query, args, err := qb.PlaceholderFormat(squirrel.Question).ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build list products query", logger.Error(err))
		return nil, false, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list products", logger.Error(err))
		return nil, false, err
	}

	defer rows.Close()

	products := make([]domain.Product, 0, limit+1)

	for rows.Next() {
		var p domain.Product
		if err := p.FromRow(rows); err != nil {
			logger.Ctx(ctx).Error("Failed to scan product row", logger.Error(err))
			return nil, false, err
		}

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore := len(products) > limit
	if hasMore {
		products = products[:limit]
	}

	return products, hasMore, nil
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
	row, err := r.queries.GetProductByBarcode(ctx, sqlc.GetProductByBarcodeParams{
		OwnerID: ownerID,
		// Match the exact string (including empty) to preserve the
		// original `barcode = ?` semantics; never coerce to NULL.
		Barcode: sql.NullString{String: barcode, Valid: true},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Product{}, domain.ErrProductNotFound
		}
		logger.Ctx(ctx).Error("Failed to look up product by barcode", logger.Error(err))
		return domain.Product{}, err
	}

	return productFromBarcodeRow(row), nil
}

func (r *Repository) CreateProduct(ctx context.Context, ownerID int64, input domain.ProductInput) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)
	result, err := qtx.CreateProduct(ctx, sqlc.CreateProductParams{
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
		BrandID:           nullInt64(input.BrandID),
		CompareAtPrice:    strconv.FormatFloat(input.CompareAtPrice, 'f', -1, 64),
		TaxRate:           strconv.FormatFloat(input.TaxRate, 'f', -1, 64),
		Currency:          defaultStr(input.Currency, "USD"),
		Weight:            strconv.FormatFloat(input.Weight, 'f', -1, 64),
		WeightUnit:        defaultStr(input.WeightUnit, "kg"),
		Length:            strconv.FormatFloat(input.Length, 'f', -1, 64),
		Width:             strconv.FormatFloat(input.Width, 'f', -1, 64),
		Height:            strconv.FormatFloat(input.Height, 'f', -1, 64),
		DimensionUnit:     defaultStr(input.DimensionUnit, "cm"),
		CountryOfOrigin:   nullString(input.CountryOfOrigin),
		HsCode:            nullString(input.HSCode),
		AllowBackorder:    input.AllowBackorder,
		LeadTimeDays:      int32(input.LeadTimeDays),
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

	if input.StockQuantity != 0 {
		if err := r.applyProductStockTotal(ctx, tx, ownerID, id, 0, input.StockQuantity, stock_core.EntryOpening, "product opening stock"); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdateProduct(ctx context.Context, ownerID, id int64, input domain.ProductInput) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	currentStock, err := r.lockProductStock(ctx, tx, ownerID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrProductNotFound
		}
		return err
	}

	qtx := r.queries.WithTx(tx)
	err = qtx.UpdateProduct(ctx, sqlc.UpdateProductParams{
		CategoryID:        nullInt64(input.CategoryID),
		SupplierID:        nullInt64(input.SupplierID),
		Name:              input.Name,
		Sku:               nullString(input.SKU),
		Barcode:           nullString(input.Barcode),
		Description:       nullString(input.Description),
		Unit:              nullString(input.Unit),
		CostPrice:         strconv.FormatFloat(input.CostPrice, 'f', -1, 64),
		SalePrice:         strconv.FormatFloat(input.SalePrice, 'f', -1, 64),
		BrandID:           nullInt64(input.BrandID),
		CompareAtPrice:    strconv.FormatFloat(input.CompareAtPrice, 'f', -1, 64),
		TaxRate:           strconv.FormatFloat(input.TaxRate, 'f', -1, 64),
		Currency:          defaultStr(input.Currency, "USD"),
		Weight:            strconv.FormatFloat(input.Weight, 'f', -1, 64),
		WeightUnit:        defaultStr(input.WeightUnit, "kg"),
		Length:            strconv.FormatFloat(input.Length, 'f', -1, 64),
		Width:             strconv.FormatFloat(input.Width, 'f', -1, 64),
		Height:            strconv.FormatFloat(input.Height, 'f', -1, 64),
		DimensionUnit:     defaultStr(input.DimensionUnit, "cm"),
		CountryOfOrigin:   nullString(input.CountryOfOrigin),
		HsCode:            nullString(input.HSCode),
		AllowBackorder:    input.AllowBackorder,
		LeadTimeDays:      int32(input.LeadTimeDays),
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

	if input.StockQuantitySet && input.StockQuantity != currentStock {
		if err := r.applyProductStockTotal(ctx, tx, ownerID, id, currentStock, input.StockQuantity, stock_core.EntryAdjustment, "product stock edit"); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *Repository) lockProductStock(ctx context.Context, tx *sql.Tx, ownerID, productID int64) (float64, error) {
	var stock string
	if err := tx.QueryRowContext(ctx,
		`SELECT stock_quantity FROM erp_products WHERE owner_id = ? AND id = ? FOR UPDATE`,
		ownerID, productID).Scan(&stock); err != nil {
		return 0, err
	}
	return parseDecimal(stock), nil
}

func (r *Repository) applyProductStockTotal(
	ctx context.Context,
	tx *sql.Tx,
	ownerID int64,
	productID int64,
	currentTotal float64,
	targetTotal float64,
	entryType stock_core.EntryType,
	note string,
) error {
	locationID, err := stock_core.DefaultLocationID(ctx, tx, ownerID)
	if err != nil {
		return err
	}

	_, _, err = stock_core.ApplyMovementTx(ctx, tx, stock_core.Movement{
		OwnerID:       ownerID,
		LocationID:    locationID,
		ProductID:     productID,
		EntryType:     entryType,
		Change:        targetTotal - currentTotal,
		ReferenceID:   int64Ptr(productID),
		ReferenceType: "product",
		Channel:       "erp",
		Note:          note,
	})
	if err != nil {
		return fmt.Errorf("sync product stock: %w", err)
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

// CountProducts returns the total number of products for the given owner,
// optionally filtered by category. Pass categoryID=0 to count all categories.
func (r *Repository) CountProducts(ctx context.Context, ownerID int64, categoryID int) (int64, error) {
	qb := squirrel.Select("COUNT(*)").
		From("erp_products").
		Where(squirrel.Eq{"owner_id": ownerID})

	if categoryID > 0 {
		qb = qb.Where(squirrel.Eq{"category_id": categoryID})
	}

	query, args, err := qb.PlaceholderFormat(squirrel.Question).ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build count products query", logger.Error(err))
		return 0, err
	}

	var count int64
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		logger.Ctx(ctx).Error("Failed to count products", logger.Error(err))
		return 0, err
	}

	return count, nil
}

func productFromRow(row sqlc.GetProductRow) domain.Product {
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

	p.CostPrice = parseDecimal(row.CostPrice)
	p.SalePrice = parseDecimal(row.SalePrice)
	p.StockQuantity = parseDecimal(row.StockQuantity)

	if row.LowStockThreshold.Valid {
		p.LowStockThreshold = parseDecimal(row.LowStockThreshold.String)
	}

	if row.ReorderQuantity.Valid {
		p.ReorderQuantity = parseDecimal(row.ReorderQuantity.String)
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

	if row.BrandID.Valid {
		p.BrandID = row.BrandID.Int64
	}
	p.CompareAtPrice = parseDecimal(row.CompareAtPrice)
	p.TaxRate = parseDecimal(row.TaxRate)
	p.Currency = row.Currency
	p.Weight = parseDecimal(row.Weight)
	p.WeightUnit = row.WeightUnit
	p.Length = parseDecimal(row.Length)
	p.Width = parseDecimal(row.Width)
	p.Height = parseDecimal(row.Height)
	p.DimensionUnit = row.DimensionUnit
	if row.CountryOfOrigin.Valid {
		p.CountryOfOrigin = row.CountryOfOrigin.String
	}
	if row.HsCode.Valid {
		p.HSCode = row.HsCode.String
	}
	p.AllowBackorder = row.AllowBackorder
	p.LeadTimeDays = int(row.LeadTimeDays)

	return p
}

func productFromBarcodeRow(row sqlc.GetProductByBarcodeRow) domain.Product {
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

	p.CostPrice = parseDecimal(row.CostPrice)
	p.SalePrice = parseDecimal(row.SalePrice)
	p.StockQuantity = parseDecimal(row.StockQuantity)

	if row.LowStockThreshold.Valid {
		p.LowStockThreshold = parseDecimal(row.LowStockThreshold.String)
	}

	if row.ReorderQuantity.Valid {
		p.ReorderQuantity = parseDecimal(row.ReorderQuantity.String)
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

	if row.BrandID.Valid {
		p.BrandID = row.BrandID.Int64
	}
	p.CompareAtPrice = parseDecimal(row.CompareAtPrice)
	p.TaxRate = parseDecimal(row.TaxRate)
	p.Currency = row.Currency
	p.Weight = parseDecimal(row.Weight)
	p.WeightUnit = row.WeightUnit
	p.Length = parseDecimal(row.Length)
	p.Width = parseDecimal(row.Width)
	p.Height = parseDecimal(row.Height)
	p.DimensionUnit = row.DimensionUnit
	if row.CountryOfOrigin.Valid {
		p.CountryOfOrigin = row.CountryOfOrigin.String
	}
	if row.HsCode.Valid {
		p.HSCode = row.HsCode.String
	}
	p.AllowBackorder = row.AllowBackorder
	p.LeadTimeDays = int(row.LeadTimeDays)

	return p
}

// ListProductAttributes returns the custom attributes for a product, ordered.
func (r *Repository) ListProductAttributes(ctx context.Context, ownerID, productID int64) ([]domain.ProductAttribute, error) {
	rows, err := r.queries.ListProductAttributes(ctx, sqlc.ListProductAttributesParams{
		OwnerID:   ownerID,
		ProductID: productID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.ProductAttribute, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.ProductAttribute{
			ID:       row.ID,
			Name:     row.Name,
			Value:    row.Value,
			Position: int(row.Position),
		})
	}
	return out, nil
}

// ReplaceProductAttributes replaces all of a product's custom attributes with
// the given set (delete-then-insert). Blank-named entries are skipped.
func (r *Repository) ReplaceProductAttributes(ctx context.Context, ownerID, productID int64, attrs []domain.ProductAttributeInput) error {
	if err := r.queries.DeleteProductAttributes(ctx, sqlc.DeleteProductAttributesParams{
		OwnerID:   ownerID,
		ProductID: productID,
	}); err != nil {
		return err
	}
	for i, a := range attrs {
		if a.Name == "" {
			continue
		}
		if err := r.queries.InsertProductAttribute(ctx, sqlc.InsertProductAttributeParams{
			OwnerID:   ownerID,
			ProductID: productID,
			Name:      a.Name,
			Value:     a.Value,
			Position:  int32(i),
		}); err != nil {
			return err
		}
	}
	return nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// defaultStr returns def when s is empty — for NOT NULL columns with a sensible
// default (currency, units) where the caller may not supply a value.
func defaultStr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func nullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}

func int64Ptr(v int64) *int64 {
	return &v
}

func nullStringFromFloat(v float64) sql.NullString {
	s := strconv.FormatFloat(v, 'f', -1, 64)
	return sql.NullString{String: s, Valid: true}
}

// parseDecimal converts a DB decimal string to float64, logging corrupt
// values instead of silently coercing them to 0.
func parseDecimal(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse product decimal from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}
