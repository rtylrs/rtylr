package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/products/internal/domain"
	"github.com/voxire/onion/internal/modules/products/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/searchutil"
	"github.com/voxire/onion/internal/platform/sqlutil"
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

// dynamic-query: squirrel
func (r *Repository) ListProducts(ctx context.Context, ownerID int64, search string, categoryID, limit, offset int) ([]domain.Product, int64, bool, error) {
	qb := squirrel.Select(
		"p.id", "p.owner_id", "p.business_id", "p.category_id", "p.supplier_id", "p.name", "p.sku",
		"p.barcode", "p.description", "p.unit", "p.cost_price", "p.sale_price",
		"p.stock_quantity", "p.low_stock_threshold", "p.reorder_quantity", "p.is_active",
		"p.track_inventory", "COALESCE(img.url, '') AS image_url", "p.created_at", "p.updated_at",
	).
		From("erp_products p").
		LeftJoin(`(
			SELECT business_id, product_id, url
			FROM (
				SELECT
					business_id,
					product_id,
					url,
					ROW_NUMBER() OVER (
						PARTITION BY business_id, product_id
						ORDER BY is_primary DESC, position ASC, id ASC
					) AS image_rank
				FROM erp_product_images
				WHERE business_id = ?
			) ranked_product_images
			WHERE image_rank = 1
		) img ON img.business_id = p.business_id AND img.product_id = p.id`, ownerID).
		Where(squirrel.Eq{"p.business_id": ownerID})

	qb = applyProductListFilters(qb, search, categoryID)
	qb = qb.OrderBy("p.name ASC").Limit(uint64(limit + 1)).Offset(uint64(offset))

	query, args, err := qb.PlaceholderFormat(squirrel.Question).ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build list products query", logger.Error(err))
		return nil, 0, false, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list products", logger.Error(err))
		return nil, 0, false, err
	}

	defer rows.Close()

	products := make([]domain.Product, 0, limit+1)

	for rows.Next() {
		var p domain.Product
		if err := p.FromRow(rows); err != nil {
			logger.Ctx(ctx).Error("Failed to scan product row", logger.Error(err))
			return nil, 0, false, err
		}

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, false, err
	}

	hasMore := len(products) > limit
	if hasMore {
		products = products[:limit]
	}

	total, err := r.countProducts(ctx, ownerID, search, categoryID)
	if err != nil {
		return nil, 0, false, err
	}

	return products, total, hasMore, nil
}

// dynamic-query: squirrel
func (r *Repository) countProducts(ctx context.Context, ownerID int64, search string, categoryID int) (int64, error) {
	qb := squirrel.Select("COUNT(*)").
		From("erp_products p").
		Where(squirrel.Eq{"p.business_id": ownerID})
	qb = applyProductListFilters(qb, search, categoryID)

	query, args, err := qb.PlaceholderFormat(squirrel.Question).ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build count products query", logger.Error(err))
		return 0, err
	}

	var total int64
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		logger.Ctx(ctx).Error("Failed to count products", logger.Error(err))
		return 0, err
	}

	return total, nil
}

// dynamic-query: squirrel
func applyProductListFilters(qb squirrel.SelectBuilder, search string, categoryID int) squirrel.SelectBuilder {
	qb = qb.Where("TRIM(COALESCE(p.name, '')) <> ''")

	if search != "" {
		booleanQuery := searchutil.BooleanPrefixQuery(search)
		if booleanQuery != "" {
			qb = qb.Where(
				"(p.sku = ? OR p.barcode = ? OR MATCH(p.name, p.sku, p.barcode) AGAINST (? IN BOOLEAN MODE))",
				search,
				search,
				booleanQuery,
			)
		} else {
			qb = qb.Where("(p.sku = ? OR p.barcode = ?)", search, search)
		}
	}

	if categoryID > 0 {
		qb = qb.Where(squirrel.Eq{"p.category_id": categoryID})
	}

	return qb
}

func (r *Repository) GetProduct(ctx context.Context, ownerID, id int64) (domain.Product, error) {
	row, err := r.queries.GetProduct(ctx, sqlc.GetProductParams{
		BusinessID: ownerID,
		ID:         id,
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
		BusinessID: ownerID,
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
	if err := validateProductReferences(ctx, qtx, ownerID, input); err != nil {
		return 0, err
	}
	result, err := qtx.CreateProduct(ctx, sqlc.CreateProductParams{
		OwnerID:           ownerID,
		BusinessID:        ownerID,
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
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return 0, domain.ErrProductAlreadyExists
		}
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

	if err := r.replacePrimaryProductImage(ctx, qtx, ownerID, id, input.Name, input.ImageURL); err != nil {
		return 0, err
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

	qtx := r.queries.WithTx(tx)

	if err := validateProductReferences(ctx, qtx, ownerID, input); err != nil {
		return err
	}

	if input.StockQuantitySet {
		locationID, err := stock_core.DefaultLocationID(ctx, tx, ownerID)
		if err != nil {
			return err
		}
		if _, _, err := stock_core.SetTotalTx(ctx, tx, stock_core.Movement{
			OwnerID:       ownerID,
			LocationID:    locationID,
			ProductID:     id,
			EntryType:     stock_core.EntryAdjustment,
			ReferenceType: "product",
			Channel:       "erp",
			Note:          "product stock edit",
		}, input.StockQuantity); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrProductNotFound
			}
			return err
		}
	}

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
		BusinessID:        ownerID,
		ID:                id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update product", logger.Error(err))

		return err
	}

	if strings.TrimSpace(input.ImageURL) != "" {
		if err := r.replacePrimaryProductImage(ctx, qtx, ownerID, id, input.Name, input.ImageURL); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func validateProductReferences(ctx context.Context, qtx *sqlc.Queries, ownerID int64, input domain.ProductInput) error {
	if input.CategoryID > 0 {
		if _, err := qtx.GetCategoryReference(ctx, sqlc.GetCategoryReferenceParams{BusinessID: ownerID, ID: input.CategoryID}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrProductReferenceNotFound
			}
			return err
		}
	}
	if input.SupplierID > 0 {
		if _, err := qtx.GetSupplierReference(ctx, sqlc.GetSupplierReferenceParams{BusinessID: ownerID, ID: input.SupplierID}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrProductReferenceNotFound
			}
			return err
		}
	}
	if input.BrandID > 0 {
		if _, err := qtx.GetBrandReference(ctx, sqlc.GetBrandReferenceParams{BusinessID: ownerID, ID: input.BrandID}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrProductReferenceNotFound
			}
			return err
		}
	}
	return nil
}

func (r *Repository) replacePrimaryProductImage(ctx context.Context, qtx *sqlc.Queries, ownerID, productID int64, alt, imageURL string) error {
	imageURL = strings.TrimSpace(imageURL)
	if err := qtx.DeletePrimaryProductImage(ctx, sqlc.DeletePrimaryProductImageParams{
		BusinessID: ownerID,
		ProductID:  productID,
	}); err != nil {
		return err
	}
	if imageURL == "" {
		return nil
	}
	return qtx.InsertPrimaryProductImage(ctx, sqlc.InsertPrimaryProductImageParams{
		OwnerID:    ownerID,
		BusinessID: ownerID,
		ProductID:  productID,
		Url:        imageURL,
		Alt:        nullString(alt),
	})
}

func (r *Repository) lockProductStock(ctx context.Context, qtx *sqlc.Queries, ownerID, productID int64) (float64, error) {
	stock, err := qtx.LockProductStockForUpdate(ctx, sqlc.LockProductStockForUpdateParams{
		BusinessID: ownerID,
		ID:         productID,
	})
	if err != nil {
		return 0, err
	}
	return sqlutil.ParseDecimal(stock), nil
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
		BusinessID: ownerID,
		ID:         id,
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
	if categoryID > 0 {
		count, err := r.queries.CountProductsByCategory(ctx, sqlc.CountProductsByCategoryParams{
			BusinessID: ownerID,
			CategoryID: sql.NullInt64{Int64: int64(categoryID), Valid: true},
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to count products by category", logger.Error(err))
			return 0, err
		}
		return count, nil
	}

	count, err := r.queries.CountAllProducts(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count products", logger.Error(err))
		return 0, err
	}
	return count, nil
}

func productFromRow(row sqlc.GetProductRow) domain.Product {
	var p domain.Product
	p.ID = row.ID
	p.OwnerID = row.OwnerID
	p.BusinessID = row.BusinessID
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

	p.CostPrice = sqlutil.ParseDecimal(row.CostPrice)
	p.SalePrice = sqlutil.ParseDecimal(row.SalePrice)
	p.StockQuantity = sqlutil.ParseDecimal(row.StockQuantity)

	if row.LowStockThreshold.Valid {
		p.LowStockThreshold = sqlutil.ParseDecimal(row.LowStockThreshold.String)
	}

	if row.ReorderQuantity.Valid {
		p.ReorderQuantity = sqlutil.ParseDecimal(row.ReorderQuantity.String)
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
	p.CompareAtPrice = sqlutil.ParseDecimal(row.CompareAtPrice)
	p.TaxRate = sqlutil.ParseDecimal(row.TaxRate)
	p.Currency = row.Currency
	p.Weight = sqlutil.ParseDecimal(row.Weight)
	p.WeightUnit = row.WeightUnit
	p.Length = sqlutil.ParseDecimal(row.Length)
	p.Width = sqlutil.ParseDecimal(row.Width)
	p.Height = sqlutil.ParseDecimal(row.Height)
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
	p.BusinessID = row.BusinessID
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

	p.CostPrice = sqlutil.ParseDecimal(row.CostPrice)
	p.SalePrice = sqlutil.ParseDecimal(row.SalePrice)
	p.StockQuantity = sqlutil.ParseDecimal(row.StockQuantity)

	if row.LowStockThreshold.Valid {
		p.LowStockThreshold = sqlutil.ParseDecimal(row.LowStockThreshold.String)
	}

	if row.ReorderQuantity.Valid {
		p.ReorderQuantity = sqlutil.ParseDecimal(row.ReorderQuantity.String)
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
	p.CompareAtPrice = sqlutil.ParseDecimal(row.CompareAtPrice)
	p.TaxRate = sqlutil.ParseDecimal(row.TaxRate)
	p.Currency = row.Currency
	p.Weight = sqlutil.ParseDecimal(row.Weight)
	p.WeightUnit = row.WeightUnit
	p.Length = sqlutil.ParseDecimal(row.Length)
	p.Width = sqlutil.ParseDecimal(row.Width)
	p.Height = sqlutil.ParseDecimal(row.Height)
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
		BusinessID: ownerID,
		ProductID:  productID,
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
		BusinessID: ownerID,
		ProductID:  productID,
	}); err != nil {
		return err
	}
	for i, a := range attrs {
		if a.Name == "" {
			continue
		}
		if err := r.queries.InsertProductAttribute(ctx, sqlc.InsertProductAttributeParams{
			OwnerID:    ownerID,
			BusinessID: ownerID,
			ProductID:  productID,
			Name:       a.Name,
			Value:      a.Value,
			Position:   int32(i),
		}); err != nil {
			return err
		}
	}
	return nil
}

var nullString = sqlutil.NullString

// defaultStr returns def when s is empty — for NOT NULL columns with a sensible
// default (currency, units) where the caller may not supply a value.
func defaultStr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

var nullInt64 = sqlutil.NullInt64

func int64Ptr(v int64) *int64 {
	return &v
}

func nullStringFromFloat(v float64) sql.NullString {
	s := strconv.FormatFloat(v, 'f', -1, 64)
	return sql.NullString{String: s, Valid: true}
}

// parseDecimal converts a DB decimal string to float64, logging corrupt
// values instead of silently coercing them to 0.
