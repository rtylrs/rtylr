package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/business/internal/domain"
	"github.com/voxire/onion/internal/modules/business/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	db      *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		db:      db,
	}
}

func (r *Repository) GetProfile(ctx context.Context, ownerID int64) (domain.BusinessProfile, error) {
	var profile domain.BusinessProfile

	if ownerID == 0 {
		return profile, errors.New("ownerID must be non-zero")
	}

	row, err := r.queries.GetBusinessProfile(ctx, int32(ownerID))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("Failed to get business profile", logger.Error(err))
		return profile, err
	}

	profile.ID = ownerID
	profile.SiteName = row.SiteName
	profile.LogoURL = row.LogoUrl
	profile.ContactEmail = row.ContactEmail
	profile.ContactPhone = row.ContactNumber
	profile.Currency = "USD"
	profile.Timezone = "UTC"
	profile.Plan = "free"

	return profile, nil
}

func (r *Repository) GetPublicBusiness(ctx context.Context, businessID int64) (domain.PublicBusiness, error) {
	// cross-schema read from businesses (shared) — public storefront lookup
	q, args, err := sq.Select(
		"id", "name", "COALESCE(slug, '')", "COALESCE(logo_url, '')",
		"COALESCE(currency, 'USD')", "COALESCE(timezone, 'UTC')",
	).
		From("businesses").
		Where(sq.Eq{"id": businessID, "is_active": true}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return domain.PublicBusiness{}, err
	}

	var res domain.PublicBusiness
	row := r.db.QueryRowContext(ctx, q, args...)
	scanErr := row.Scan(&res.ID, &res.Name, &res.Slug, &res.LogoURL, &res.Currency, &res.Timezone)
	return res, scanErr
}

func (r *Repository) LookupBySlug(ctx context.Context, slug string) (domain.SlugLookupResult, error) {
	var out domain.SlugLookupResult

	// cross-schema read from businesses (shared) — resolve storefront by slug
	bizQ, bizArgs, err := sq.Select(
		"id", "name", "COALESCE(slug, '')", "COALESCE(logo_url, '')",
		"COALESCE(currency, 'USD')", "COALESCE(timezone, 'UTC')", "COALESCE(owner_id, 0)",
	).
		From("businesses").
		Where(sq.Eq{"slug": slug, "is_active": true}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return out, err
	}

	var ownerID int64
	row := r.db.QueryRowContext(ctx, bizQ, bizArgs...)
	if err := row.Scan(&out.Business.ID, &out.Business.Name, &out.Business.Slug, &out.Business.LogoURL, &out.Business.Currency, &out.Business.Timezone, &ownerID); err != nil {
		return out, err
	}

	// cross-schema read from hr_locations (hr_core) — active storefront locations
	locQ, locArgs, err := sq.Select(
		"id", "name", "COALESCE(slug, '')", "COALESCE(address, '')", "COALESCE(timezone, '')", "is_active",
	).
		From("hr_locations").
		Where(sq.Eq{"owner_id": ownerID, "is_active": true}).
		OrderBy("is_default DESC", "name ASC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return out, err
	}

	rows, err := r.db.QueryContext(ctx, locQ, locArgs...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return out, nil
		}
		return out, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var loc domain.PublicLocation
		if err := rows.Scan(&loc.ID, &loc.Name, &loc.Slug, &loc.Address, &loc.Timezone, &loc.IsActive); err != nil {
			return out, err
		}
		out.Locations = append(out.Locations, loc)
	}
	return out, rows.Err()
}

func (r *Repository) UpdateSlug(ctx context.Context, businessID int64, slug string) error {
	// cross-schema write to businesses (shared) — set storefront slug
	q, args, err := sq.Update("businesses").
		Set("slug", slug).
		Where(sq.Eq{"id": businessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, q, args...)
	return err
}
