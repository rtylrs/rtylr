package infrastructure

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/voxire/onion/internal/modules/store_products/internal/domain"
	"github.com/voxire/onion/internal/modules/store_products/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{queries: sqlc.New(pool), pool: pool}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.StoreProduct, error) {
	rows, err := r.queries.ListStoreProducts(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list store products", logger.Error(err))
		return nil, err
	}
	items := make([]domain.StoreProduct, 0, len(rows))
	for _, row := range rows {
		items = append(items, fromRow(ctx, row))
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.StoreProductInput) (int64, error) {
	result, err := r.queries.CreateStoreProduct(ctx, sqlc.CreateStoreProductParams{
		OwnerID:        ownerID,
		ProductID:      sql.NullInt64{Int64: input.ProductID, Valid: input.ProductID != 0},
		Title:          input.Title,
		Slug:           sql.NullString{String: input.Slug, Valid: input.Slug != ""},
		Description:    sql.NullString{String: input.Description, Valid: input.Description != ""},
		Images:         nullJSON(input.Images),
		Price:          strconv.FormatFloat(input.Price, 'f', 2, 64),
		CompareAtPrice: nullDecimal(input.CompareAtPrice),
		IsPublished:    sql.NullBool{Bool: input.IsPublished, Valid: true},
		SortOrder:      sql.NullInt32{Int32: int32(input.SortOrder), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create store product", logger.Error(err))
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func (r *Repository) Update(ctx context.Context, ownerID, id int64, input domain.StoreProductInput) error {
	err := r.queries.UpdateStoreProduct(ctx, sqlc.UpdateStoreProductParams{
		ProductID:      sql.NullInt64{Int64: input.ProductID, Valid: input.ProductID != 0},
		Title:          input.Title,
		Slug:           sql.NullString{String: input.Slug, Valid: input.Slug != ""},
		Description:    sql.NullString{String: input.Description, Valid: input.Description != ""},
		Images:         nullJSON(input.Images),
		Price:          strconv.FormatFloat(input.Price, 'f', 2, 64),
		CompareAtPrice: nullDecimal(input.CompareAtPrice),
		IsPublished:    sql.NullBool{Bool: input.IsPublished, Valid: true},
		SortOrder:      sql.NullInt32{Int32: int32(input.SortOrder), Valid: true},
		OwnerID:        ownerID,
		ID:             id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update store product", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, ownerID, id int64) error {
	err := r.queries.DeleteStoreProduct(ctx, sqlc.DeleteStoreProductParams{OwnerID: ownerID, ID: id})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete store product", logger.Error(err))
		return err
	}
	return nil
}

func parseMonetary(ctx context.Context, s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Ctx(ctx).Warn("Failed to parse monetary value", logger.String("value", s), logger.Error(err))
		return 0
	}
	return v
}

func fromRow(ctx context.Context, row sqlc.StoreProduct) domain.StoreProduct {
	p := domain.StoreProduct{ID: row.ID, OwnerID: row.OwnerID, Title: row.Title}
	if row.ProductID.Valid {
		p.ProductID = row.ProductID.Int64
	}
	if row.Slug.Valid {
		p.Slug = row.Slug.String
	}
	if row.Description.Valid {
		p.Description = row.Description.String
	}
	if row.Images != nil {
		p.Images = string(row.Images)
	}
	p.Price = parseMonetary(ctx, row.Price)
	if row.CompareAtPrice.Valid {
		p.CompareAtPrice = parseMonetary(ctx, row.CompareAtPrice.String)
	}
	if row.IsPublished.Valid {
		p.IsPublished = row.IsPublished.Bool
	}
	if row.SortOrder.Valid {
		p.SortOrder = int(row.SortOrder.Int32)
	}
	if row.CreatedAt.Valid {
		p.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		p.UpdatedAt = row.UpdatedAt.Time
	}
	return p
}

func nullJSON(s string) []byte {
	if s == "" {
		return nil
	}
	return []byte(s)
}

func nullDecimal(v float64) sql.NullString {
	return sql.NullString{String: strconv.FormatFloat(v, 'f', 2, 64), Valid: true}
}
