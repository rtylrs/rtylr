package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/voxire/onion/internal/modules/variants/internal/domain"
	"github.com/voxire/onion/internal/modules/variants/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) ListVariants(ctx context.Context, ownerID, productID int64) ([]domain.Variant, error) {
	var rows []sqlc.ErpProductVariant
	var err error

	if productID > 0 {
		rows, err = r.queries.ListVariantsByProduct(ctx, sqlc.ListVariantsByProductParams{
			OwnerID:   ownerID,
			ProductID: productID,
		})
	} else {
		rows, err = r.queries.ListVariantsByOwner(ctx, ownerID)
	}

	if err != nil {
		logger.Ctx(ctx).Error("Failed to list variants", logger.Error(err))

		return nil, err
	}

	var variants []domain.Variant

	for _, row := range rows {
		variants = append(variants, variantFromRow(row))
	}

	return variants, nil
}

func (r *Repository) CreateVariant(ctx context.Context, ownerID int64, input domain.VariantInput) (int64, error) {
	result, err := r.queries.CreateVariant(ctx, sqlc.CreateVariantParams{
		OwnerID:       ownerID,
		ProductID:     input.ProductID,
		Name:          input.Name,
		Sku:           nullString(input.SKU),
		Barcode:       nullString(input.Barcode),
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

func (r *Repository) UpdateVariant(ctx context.Context, ownerID, id int64, input domain.VariantInput) error {
	err := r.queries.UpdateVariant(ctx, sqlc.UpdateVariantParams{
		ProductID:     input.ProductID,
		Name:          input.Name,
		Sku:           nullString(input.SKU),
		Barcode:       nullString(input.Barcode),
		CostPrice:     strconv.FormatFloat(input.CostPrice, 'f', -1, 64),
		SalePrice:     strconv.FormatFloat(input.SalePrice, 'f', -1, 64),
		StockQuantity: strconv.FormatFloat(input.StockQuantity, 'f', -1, 64),
		Attributes:    nullJSON(input.Attributes),
		IsActive:      sql.NullBool{Bool: input.IsActive, Valid: true},
		OwnerID:       ownerID,
		ID:            id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update variant", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) DeleteVariant(ctx context.Context, ownerID, id int64) error {
	err := r.queries.DeleteVariant(ctx, sqlc.DeleteVariantParams{
		OwnerID: ownerID,
		ID:      id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete variant", logger.Error(err))

		return err
	}

	return nil
}

func variantFromRow(row sqlc.ErpProductVariant) domain.Variant {
	var v domain.Variant
	v.ID = row.ID
	v.OwnerID = row.OwnerID
	v.ProductID = row.ProductID
	v.Name = row.Name

	if row.Sku.Valid {
		v.SKU = row.Sku.String
	}

	if row.Barcode.Valid {
		v.Barcode = row.Barcode.String
	}

	v.CostPrice, _ = strconv.ParseFloat(row.CostPrice, 64)
	v.SalePrice, _ = strconv.ParseFloat(row.SalePrice, 64)
	v.StockQuantity, _ = strconv.ParseFloat(row.StockQuantity, 64)

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

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullJSON(s string) json.RawMessage {
	if s == "" {
		return nil
	}

	return json.RawMessage(s)
}
