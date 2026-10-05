package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/erp/internal/outbox"

	"github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/erp/internal/stock"
	"github.com/voxire/onion/internal/modules/erp/products/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/products/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/searchutil"
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
	qb = qb.OrderBy("p.name ASC").Limit(uint64(limit + 1)).Offset(uint64(offset)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.

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
		p, err := productFromSQLRow(rows)
		if err != nil {
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
		LeadTimeDays:      int32(input.LeadTimeDays), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
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
		if err := r.applyProductStockTotal(ctx, tx, ownerID, id, 0, input.StockQuantity, stock.EntryOpening, "product opening stock"); err != nil {
			return 0, err
		}
	}

	if err := r.replacePrimaryProductImage(ctx, qtx, ownerID, id, input.Name, input.ImageURL); err != nil {
		return 0, err
	}

	if err := outbox.AppendAuditTx(ctx, tx, audit.Event{BusinessID: ownerID, EntityType: "product", EntityID: id, Action: "created", NewValues: productInputSnapshot(input)}); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) SalesProduct(ctx context.Context, businessID, productID int64) (name, sku, price, tax string, err error) {
	row, err := r.queries.GetSalesProduct(ctx, sqlc.GetSalesProductParams{BusinessID: businessID, ID: productID})
	if err != nil {
		return "", "", "", "", err
	}
	return row.Name, row.Sku, row.SalePrice, row.TaxRate, nil
}
