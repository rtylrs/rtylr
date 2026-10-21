package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/voxire/onion/internal/modules/erp/variants/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/variants/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
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

func (r *Repository) ListVariants(ctx context.Context, businessID, productID int64) ([]domain.Variant, error) {
	items, _, _, err := r.ListVariantsPage(ctx, businessID, productID, 100, 0)
	return items, err
}

func (r *Repository) ListVariantsPage(ctx context.Context, businessID, productID int64, limit, offset int) ([]domain.Variant, bool, int64, error) {
	if productID > 0 {
		if err := ensureProductReference(ctx, r.queries, businessID, productID); err != nil {
			return nil, false, 0, err
		}
	}

	var rows []sqlc.ErpProductVariant
	var err error
	if productID > 0 {
		rows, err = r.queries.ListVariantsByProduct(ctx, sqlc.ListVariantsByProductParams{
			BusinessID: businessID,
			ProductID:  productID,
			Limit:      int32(limit + 1), // #nosec G115 -- limit is validated by the application service.
			Offset:     int32(offset),    // #nosec G115 -- offset is derived from validated service bounds.
		})
	} else {
		rows, err = r.queries.ListVariantsByOwner(ctx, sqlc.ListVariantsByOwnerParams{
			BusinessID: businessID,
			Limit:      int32(limit + 1), // #nosec G115 -- limit is validated by the application service.
			Offset:     int32(offset),    // #nosec G115 -- offset is derived from validated service bounds.
		})
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list variants", logger.Error(err))
		return nil, false, 0, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	total, err := r.CountVariants(ctx, businessID, productID)
	if err != nil {
		return nil, false, 0, err
	}
	variants := make([]domain.Variant, 0, len(rows))
	for _, row := range rows {
		variants = append(variants, variantFromRow(row))
	}
	return variants, hasMore, total, nil
}

func (r *Repository) CreateVariant(ctx context.Context, businessID int64, input domain.VariantInput) (id int64, err error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	qtx := r.queries.WithTx(tx)
	if err = ensureProductReference(ctx, qtx, businessID, input.ProductID); err != nil {
		return 0, err
	}
	result, err := qtx.CreateVariant(ctx, sqlc.CreateVariantParams{
		OwnerID:       businessID,
		BusinessID:    businessID,
		ProductID:     input.ProductID,
		Name:          input.Name,
		Sku:           sqlutil.NullString(input.SKU),
		Barcode:       sqlutil.NullString(input.Barcode),
		CostPrice:     strconv.FormatFloat(input.CostPrice, 'f', -1, 64),
		SalePrice:     strconv.FormatFloat(input.SalePrice, 'f', -1, 64),
		StockQuantity: strconv.FormatFloat(input.StockQuantity, 'f', -1, 64),
		Attributes:    nullJSON(input.Attributes),
		IsActive:      sql.NullBool{Bool: input.IsActive, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create variant", logger.Error(err))
		return 0, err
	}
	id, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}
	err = tx.Commit()
	return id, err
}

func (r *Repository) UpdateVariant(ctx context.Context, businessID, id int64, input domain.VariantInput) (err error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	qtx := r.queries.WithTx(tx)
	if err = ensureProductReference(ctx, qtx, businessID, input.ProductID); err != nil {
		return err
	}
	updated, err := qtx.UpdateVariant(ctx, sqlc.UpdateVariantParams{
		ProductID:     input.ProductID,
		Name:          input.Name,
		Sku:           sqlutil.NullString(input.SKU),
		Barcode:       sqlutil.NullString(input.Barcode),
		CostPrice:     strconv.FormatFloat(input.CostPrice, 'f', -1, 64),
		SalePrice:     strconv.FormatFloat(input.SalePrice, 'f', -1, 64),
		StockQuantity: strconv.FormatFloat(input.StockQuantity, 'f', -1, 64),
		Attributes:    nullJSON(input.Attributes),
		IsActive:      sql.NullBool{Bool: input.IsActive, Valid: true},
		BusinessID:    businessID,
		ID:            id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update variant", logger.Error(err))
		return err
	}
	if updated != 1 {
		return domain.ErrVariantNotFound
	}
	err = tx.Commit()
	return err
}

func (r *Repository) DeleteVariant(ctx context.Context, businessID, id int64) error {
	deleted, err := r.queries.DeleteVariant(ctx, sqlc.DeleteVariantParams{BusinessID: businessID, ID: id})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete variant", logger.Error(err))
		return err
	}
	if deleted != 1 {
		return domain.ErrVariantNotFound
	}
	return nil
}

// CountVariants returns the number of variants for a given owner.
// Pass productID=0 to count across all products, or a positive productID to
// limit the count to that product.
func (r *Repository) CountVariants(ctx context.Context, businessID, productID int64) (int64, error) {
	var (
		count int64
		err   error
	)

	if productID > 0 {
		count, err = r.queries.CountVariantsByProduct(ctx, sqlc.CountVariantsByProductParams{
			BusinessID: businessID,
			ProductID:  productID,
		})
	} else {
		count, err = r.queries.CountVariantsByOwner(ctx, businessID)
	}

	if err != nil {
		logger.Ctx(ctx).Error("Failed to count variants", logger.Error(err))
		return 0, err
	}

	return count, nil
}

func ensureProductReference(ctx context.Context, q *sqlc.Queries, businessID, productID int64) error {
	if productID <= 0 {
		return domain.ErrVariantProductNotFound
	}
	if _, err := q.GetProductReference(ctx, sqlc.GetProductReferenceParams{BusinessID: businessID, ID: productID}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrVariantProductNotFound
		}
		return err
	}
	return nil
}

func variantFromRow(row sqlc.ErpProductVariant) domain.Variant {
	var v domain.Variant
	v.ID = row.ID
	v.OwnerID = row.OwnerID
	v.BusinessID = row.BusinessID
	v.ProductID = row.ProductID
	v.Name = row.Name

	if row.Sku.Valid {
		v.SKU = row.Sku.String
	}

	if row.Barcode.Valid {
		v.Barcode = row.Barcode.String
	}

	v.CostPrice = sqlutil.ParseDecimal(row.CostPrice)
	v.SalePrice = sqlutil.ParseDecimal(row.SalePrice)
	v.StockQuantity = sqlutil.ParseDecimal(row.StockQuantity)

	if row.Attributes != nil {
		v.Attributes = string(row.Attributes)
	}

	if row.IsActive.Valid {
		v.IsActive = row.IsActive.Bool
	}

	if row.CreatedAt.Valid {
		v.CreatedAt = row.CreatedAt.Time
	}

	if row.UpdatedAt.Valid {
		v.UpdatedAt = row.UpdatedAt.Time
	}

	return v
}

func nullJSON(s string) json.RawMessage {
	if s == "" {
		return nil
	}

	return json.RawMessage(s)
}

// parseDecimal converts a DB decimal string to float64, logging corrupt
// values instead of silently coercing them to 0.
