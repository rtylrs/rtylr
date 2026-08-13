package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/brands/internal/domain"
	"github.com/voxire/onion/internal/modules/brands/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) ListBrands(ctx context.Context, businessID int64) ([]domain.Brand, error) {
	rows, err := r.queries.ListBrands(ctx, sqlc.ListBrandsParams{
		BusinessID:   businessID,
		BusinessID_2: businessID,
	})
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
		BusinessID:   businessID,
		BusinessID_2: businessID,
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list brands page", logger.Error(err))

		return nil, err
	}

	brands := make([]domain.Brand, 0, len(rows))

	for _, row := range rows {
		brands = append(brands, brandFromRow(sqlc.ListBrandsRow(row)))
	}

	return brands, nil
}

// brandFromRow maps a sqlc erp_brands row (with its computed product count)
// to the domain model, collapsing nullable columns to their zero value when
// NULL.
func brandFromRow(row sqlc.ListBrandsRow) domain.Brand {
	var b domain.Brand
	b.ID = row.ID
	b.OwnerID = row.OwnerID
	b.BusinessID = row.BusinessID
	b.Name = row.Name
	b.ProductCount = row.ProductCount

	if row.LogoUrl.Valid {
		b.LogoURL = row.LogoUrl.String
	}

	if row.Description.Valid {
		b.Description = row.Description.String
	}

	if row.CreatedAt.Valid {
		b.CreatedAt = row.CreatedAt.Time
	}

	if row.UpdatedAt.Valid {
		b.UpdatedAt = row.UpdatedAt.Time
	}

	return b
}

func (r *Repository) CreateBrand(ctx context.Context, businessID int64, input domain.BrandInput) (int64, error) {
	result, err := r.queries.CreateBrand(ctx, sqlc.CreateBrandParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		Name:        input.Name,
		LogoUrl:     nullString(input.LogoURL),
		Description: nullString(input.Description),
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
		LogoUrl:     nullString(input.LogoURL),
		Description: nullString(input.Description),
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

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
