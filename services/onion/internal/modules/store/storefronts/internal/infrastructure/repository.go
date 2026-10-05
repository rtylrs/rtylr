package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/store/storefronts/internal/domain"
	"github.com/voxire/onion/internal/modules/store/storefronts/internal/infrastructure/persistence/sqlc"
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

// ThemeBelongsToOwner reports whether the given theme exists and belongs to businessID.
func (r *Repository) ThemeBelongsToOwner(ctx context.Context, businessID, themeID int64) (bool, error) {
	// cross-schema read from store_themes — verify theme ownership before linking a storefront
	count, err := r.queries.CountThemeForOwner(ctx, sqlc.CountThemeForOwnerParams{ID: themeID, BusinessID: businessID})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to verify theme ownership", logger.Error(err))
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.Storefront, error) {
	rows, err := r.queries.ListStorefronts(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list storefronts", logger.Error(err))
		return nil, err
	}
	items := make([]domain.Storefront, 0, len(rows))
	for _, row := range rows {
		items = append(items, fromRow(row))
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, creatorID, businessID int64, input domain.StorefrontInput) (int64, error) {
	result, err := r.queries.CreateStorefront(ctx, sqlc.CreateStorefrontParams{
		OwnerID:     creatorID,
		BusinessID:  businessID,
		Name:        input.Name,
		Domain:      sql.NullString{String: input.Domain, Valid: input.Domain != ""},
		LogoUrl:     sql.NullString{String: input.LogoURL, Valid: input.LogoURL != ""},
		ThemeID:     sql.NullInt64{Int64: input.ThemeID, Valid: input.ThemeID != 0},
		IsPublished: sql.NullBool{Bool: input.IsPublished, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create storefront", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.StorefrontInput) error {
	err := r.queries.UpdateStorefront(ctx, sqlc.UpdateStorefrontParams{
		Name:        input.Name,
		Domain:      sql.NullString{String: input.Domain, Valid: input.Domain != ""},
		LogoUrl:     sql.NullString{String: input.LogoURL, Valid: input.LogoURL != ""},
		ThemeID:     sql.NullInt64{Int64: input.ThemeID, Valid: input.ThemeID != 0},
		IsPublished: sql.NullBool{Bool: input.IsPublished, Valid: true},
		BusinessID:  businessID,
		ID:          id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update storefront", logger.Error(err))
		return err
	}
	return nil
}

func fromRow(row sqlc.StoreSite) domain.Storefront {
	s := domain.Storefront{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		BusinessID: row.BusinessID,
		Name:       row.Name,
	}
	if row.Domain.Valid {
		s.Domain = row.Domain.String
	}
	if row.LogoUrl.Valid {
		s.LogoURL = row.LogoUrl.String
	}
	if row.ThemeID.Valid {
		s.ThemeID = row.ThemeID.Int64
	}
	if row.IsPublished.Valid {
		s.IsPublished = row.IsPublished.Bool
	}
	if row.CreatedAt.Valid {
		s.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		s.UpdatedAt = row.UpdatedAt.Time
	}
	return s
}
