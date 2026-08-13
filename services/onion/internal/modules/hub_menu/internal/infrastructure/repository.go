package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/hub_menu/internal/domain"
	"github.com/voxire/onion/internal/modules/hub_menu/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{queries: sqlc.New(db)}
}

func (r *Repository) ListBanners(ctx context.Context, businessID int64) ([]domain.Banner, error) {
	rows, err := r.queries.ListMenuBanners(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list hub menu banners", logger.Error(err))
		return nil, err
	}
	banners := make([]domain.Banner, 0, len(rows))
	for _, row := range rows {
		banners = append(banners, listRowToBanner(row))
	}
	return banners, nil
}

func (r *Repository) CreateBanner(ctx context.Context, ownerID, businessID int64, input domain.BannerInput) (int64, error) {
	result, err := r.queries.CreateMenuBanner(ctx, sqlc.CreateMenuBannerParams{
		OwnerID:    ownerID,
		BusinessID: businessID,
		Title:      input.Title,
		Body:       nullString(input.Body),
		ImageUrl:   nullString(input.ImageURL),
		LinkUrl:    nullString(input.LinkURL),
		Position:   input.Position,
		IsActive:   input.IsActive,
		StartsAt:   nullTime(input.StartsAt),
		EndsAt:     nullTime(input.EndsAt),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create hub menu banner", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) UpdateBanner(ctx context.Context, businessID, id int64, input domain.BannerInput) error {
	if err := r.queries.UpdateMenuBanner(ctx, sqlc.UpdateMenuBannerParams{
		Title:      input.Title,
		Body:       nullString(input.Body),
		ImageUrl:   nullString(input.ImageURL),
		LinkUrl:    nullString(input.LinkURL),
		Position:   input.Position,
		IsActive:   input.IsActive,
		StartsAt:   nullTime(input.StartsAt),
		EndsAt:     nullTime(input.EndsAt),
		BusinessID: businessID,
		ID:         id,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to update hub menu banner", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) DeleteBanner(ctx context.Context, businessID, id int64) error {
	if err := r.queries.DeleteMenuBanner(ctx, sqlc.DeleteMenuBannerParams{
		BusinessID: businessID,
		ID:         id,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to delete hub menu banner", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) GetBanner(ctx context.Context, businessID, id int64) (domain.Banner, error) {
	row, err := r.queries.GetMenuBanner(ctx, sqlc.GetMenuBannerParams{
		BusinessID: businessID,
		ID:         id,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Banner{}, domain.ErrBannerNotFound
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to load hub menu banner", logger.Error(err))
		return domain.Banner{}, err
	}
	return getRowToBanner(row), nil
}

func (r *Repository) BusinessOwnerID(ctx context.Context, businessID int64) (int64, error) {
	ownerID, err := r.queries.GetHubMenuBusinessOwnerID(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to resolve hub menu business owner", logger.Error(err))
		return 0, err
	}
	return ownerID, nil
}

func listRowToBanner(row sqlc.ListMenuBannersRow) domain.Banner {
	return bannerFromFields(domain.Banner{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		BusinessID: row.BusinessID,
		Title:      row.Title,
		Body:       row.Body,
		ImageURL:   row.ImageUrl,
		LinkURL:    row.LinkUrl,
		Position:   row.Position,
		IsActive:   row.IsActive,
	}, row.StartsAt, row.EndsAt, row.CreatedAt, row.UpdatedAt)
}

func getRowToBanner(row sqlc.GetMenuBannerRow) domain.Banner {
	return bannerFromFields(domain.Banner{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		BusinessID: row.BusinessID,
		Title:      row.Title,
		Body:       row.Body,
		ImageURL:   row.ImageUrl,
		LinkURL:    row.LinkUrl,
		Position:   row.Position,
		IsActive:   row.IsActive,
	}, row.StartsAt, row.EndsAt, row.CreatedAt, row.UpdatedAt)
}

func bannerFromFields(
	banner domain.Banner,
	startsAt sql.NullTime,
	endsAt sql.NullTime,
	createdAt sql.NullTime,
	updatedAt sql.NullTime,
) domain.Banner {
	if startsAt.Valid {
		banner.StartsAt = startsAt.Time
	}
	if endsAt.Valid {
		banner.EndsAt = endsAt.Time
	}
	if createdAt.Valid {
		banner.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		banner.UpdatedAt = updatedAt.Time
	}
	return banner
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func nullTime(value time.Time) sql.NullTime {
	if value.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: value, Valid: true}
}
