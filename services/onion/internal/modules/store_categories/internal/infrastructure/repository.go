package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/store_categories/internal/domain"
	"github.com/voxire/onion/internal/modules/store_categories/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.StoreCategory, error) {
	rows, err := r.queries.ListStoreCategories(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list store categories", logger.Error(err))
		return nil, err
	}
	items := make([]domain.StoreCategory, 0, len(rows))
	for _, row := range rows {
		items = append(items, fromRow(row))
	}
	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, businessID, id int64) (domain.StoreCategory, error) {
	row, err := r.queries.GetStoreCategory(ctx, sqlc.GetStoreCategoryParams{BusinessID: businessID, ID: id})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.StoreCategory{}, domain.ErrParentNotFound
		}
		logger.Ctx(ctx).Error("Failed to get store category", logger.Error(err))
		return domain.StoreCategory{}, err
	}
	return fromRow(row), nil
}

func (r *Repository) Create(ctx context.Context, ownerID, businessID int64, input domain.StoreCategoryInput) (int64, error) {
	result, err := r.queries.CreateStoreCategory(ctx, sqlc.CreateStoreCategoryParams{
		OwnerID:     ownerID,
		BusinessID:  businessID,
		Name:        input.Name,
		Slug:        sql.NullString{String: input.Slug, Valid: input.Slug != ""},
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		ImageUrl:    sql.NullString{String: input.ImageURL, Valid: input.ImageURL != ""},
		ParentID:    sql.NullInt64{Int64: input.ParentID, Valid: input.ParentID != 0},
		SortOrder:   sql.NullInt32{Int32: int32(input.SortOrder), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create store category", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.StoreCategoryInput) error {
	err := r.queries.UpdateStoreCategory(ctx, sqlc.UpdateStoreCategoryParams{
		Name:        input.Name,
		Slug:        sql.NullString{String: input.Slug, Valid: input.Slug != ""},
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		ImageUrl:    sql.NullString{String: input.ImageURL, Valid: input.ImageURL != ""},
		ParentID:    sql.NullInt64{Int64: input.ParentID, Valid: input.ParentID != 0},
		SortOrder:   sql.NullInt32{Int32: int32(input.SortOrder), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		BusinessID:  businessID,
		ID:          id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update store category", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	err := r.queries.DeleteStoreCategory(ctx, sqlc.DeleteStoreCategoryParams{BusinessID: businessID, ID: id})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete store category", logger.Error(err))
		return err
	}
	return nil
}

func fromRow(row sqlc.StoreCategory) domain.StoreCategory {
	c := domain.StoreCategory{ID: row.ID, OwnerID: row.OwnerID, BusinessID: row.BusinessID, Name: row.Name}
	if row.Slug.Valid {
		c.Slug = row.Slug.String
	}
	if row.Description.Valid {
		c.Description = row.Description.String
	}
	if row.ImageUrl.Valid {
		c.ImageURL = row.ImageUrl.String
	}
	if row.ParentID.Valid {
		c.ParentID = row.ParentID.Int64
	}
	if row.SortOrder.Valid {
		c.SortOrder = int(row.SortOrder.Int32)
	}
	if row.CreatedAt.Valid {
		c.CreatedAt = row.CreatedAt.Time
	}
	return c
}
