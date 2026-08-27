package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/voxire/onion/internal/modules/variants/internal/domain"
	"github.com/voxire/onion/internal/modules/variants/internal/infrastructure/persistence/sqlc"
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
	var rows []sqlc.ErpProductVariant
	var err error

	if productID > 0 {
		rows, err = r.queries.ListVariantsByProduct(ctx, sqlc.ListVariantsByProductParams{
			BusinessID: businessID,
			ProductID:  productID,
		})
	} else {
		rows, err = r.queries.ListVariantsByOwner(ctx, businessID)
	}

	if err != nil {
		logger.Ctx(ctx).Error("Failed to list variants", logger.Error(err))

		return nil, err
	}

	variants := make([]domain.Variant, 0, len(rows))

	for _, row := range rows {
		variants = append(variants, variantFromRow(row))
	}

	return variants, nil
}

func (r *Repository) CreateVariant(ctx context.Context, businessID int64, input domain.VariantInput) (int64, error) {
	result, err := r.queries.CreateVariant(ctx, sqlc.CreateVariantParams{
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

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdateVariant(ctx context.Context, businessID, id int64, input domain.VariantInput) error {
	err := r.queries.UpdateVariant(ctx, sqlc.UpdateVariantParams{
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

	return nil
}

func (r *Repository) DeleteVariant(ctx context.Context, businessID, id int64) error {
	err := r.queries.DeleteVariant(ctx, sqlc.DeleteVariantParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete variant", logger.Error(err))

		return err
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
