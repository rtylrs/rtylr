package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/brands/internal/domain"
	"github.com/voxire/onion/internal/modules/brands/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) ListBrands(ctx context.Context, ownerID int64) ([]domain.Brand, error) {
	rows, err := r.queries.ListBrands(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list brands", logger.Error(err))

		return nil, err
	}

	var brands []domain.Brand

	for _, row := range rows {
		var b domain.Brand
		b.ID = row.ID
		b.OwnerID = row.OwnerID
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

		brands = append(brands, b)
	}

	return brands, nil
}

func (r *Repository) CreateBrand(ctx context.Context, ownerID int64, input domain.BrandInput) (int64, error) {
	result, err := r.queries.CreateBrand(ctx, sqlc.CreateBrandParams{
		OwnerID:     ownerID,
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

func (r *Repository) UpdateBrand(ctx context.Context, ownerID, id int64, input domain.BrandInput) error {
	err := r.queries.UpdateBrand(ctx, sqlc.UpdateBrandParams{
		Name:        input.Name,
		LogoUrl:     nullString(input.LogoURL),
		Description: nullString(input.Description),
		OwnerID:     ownerID,
		ID:          id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update brand", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) DeleteBrand(ctx context.Context, ownerID, id int64) error {
	err := r.queries.DeleteBrand(ctx, sqlc.DeleteBrandParams{
		OwnerID: ownerID,
		ID:      id,
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
