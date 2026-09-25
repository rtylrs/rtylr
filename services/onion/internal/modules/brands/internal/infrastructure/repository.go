package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/brands/internal/domain"
	"github.com/voxire/onion/internal/modules/brands/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) ListBrands(ctx context.Context, businessID int64) ([]domain.Brand, error) {
	rows, err := r.queries.ListBrands(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list brands", logger.Error(err))

		return nil, err
	}

	brands := make([]domain.Brand, 0, len(rows))

	for _, row := range rows {
		brands = append(brands, brandFromRow(row))
	}

	return brands, nil
}

// ListBrandsPaginated returns a single page of brands for the business,
// ordered by name. Limit and offset are applied as-is; callers clamp them.
func (r *Repository) ListBrandsPaginated(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Brand, error) {
	rows, err := r.queries.ListBrandsPaginated(ctx, sqlc.ListBrandsPaginatedParams{
		BusinessID: businessID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list brands page", logger.Error(err))

		return nil, err
	}

	brands := make([]domain.Brand, 0, len(rows))

	for _, row := range rows {
		brands = append(brands, brandFromPaginatedRow(row))
	}

	return brands, nil
}

func (r *Repository) CountBrands(ctx context.Context, businessID int64) (int64, error) {
	total, err := r.queries.CountBrands(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count brands", logger.Error(err))
		return 0, err
	}
	return total, nil
}

// brandFromRow maps a sqlc erp_brands row (with its computed product count)
// to the domain model, collapsing nullable columns to their zero value when
// NULL.
func brandFromRow(row sqlc.ListBrandsRow) domain.Brand {
	return brandFromFields(
		row.ID,
		row.OwnerID,
		row.BusinessID,
		row.Name,
		row.LogoUrl,
		row.Description,
		row.ProductCount,
		row.CreatedAt,
		row.UpdatedAt,
	)
}

func brandFromPaginatedRow(row sqlc.ListBrandsPaginatedRow) domain.Brand {
	return brandFromFields(
		row.ID,
		row.OwnerID,
		row.BusinessID,
		row.Name,
		row.LogoUrl,
		row.Description,
		row.ProductCount,
		row.CreatedAt,
		row.UpdatedAt,
	)
}

func brandFromFields(
	id int64,
	ownerID int64,
	businessID int64,
	name string,
	logoURL sql.NullString,
	description sql.NullString,
	productCount int64,
	createdAt sql.NullTime,
	updatedAt sql.NullTime,
) domain.Brand {
	var b domain.Brand
	b.ID = id
	b.OwnerID = ownerID
	b.BusinessID = businessID
	b.Name = name
	b.ProductCount = productCount

	if logoURL.Valid {
		b.LogoURL = logoURL.String
	}

	if description.Valid {
		b.Description = description.String
	}

	if createdAt.Valid {
		b.CreatedAt = createdAt.Time
	}

	if updatedAt.Valid {
		b.UpdatedAt = updatedAt.Time
	}

	return b
}

func (r *Repository) CreateBrand(ctx context.Context, businessID int64, input domain.BrandInput) (int64, error) {
	result, err := r.queries.CreateBrand(ctx, sqlc.CreateBrandParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		Name:        input.Name,
		LogoUrl:     sqlutil.NullString(input.LogoURL),
		Description: sqlutil.NullString(input.Description),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create brand", logger.Error(err))

		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdateBrand(ctx context.Context, businessID, id int64, input domain.BrandInput) error {
	err := r.queries.UpdateBrand(ctx, sqlc.UpdateBrandParams{
		Name:        input.Name,
		LogoUrl:     sqlutil.NullString(input.LogoURL),
		Description: sqlutil.NullString(input.Description),
		BusinessID:  businessID,
		ID:          id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update brand", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) DeleteBrand(ctx context.Context, businessID, id int64) error {
	err := r.queries.DeleteBrand(ctx, sqlc.DeleteBrandParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete brand", logger.Error(err))

		return err
	}

	return nil
}
