package infrastructure

import (
	"context"
	"database/sql"
	"errors"

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
	const q = `SELECT id, name, COALESCE(slug, ''), COALESCE(logo_url, ''), COALESCE(currency, 'USD'), COALESCE(timezone, 'UTC')
	           FROM businesses WHERE id = ? AND is_active = TRUE LIMIT 1`

	var res domain.PublicBusiness
	row := r.db.QueryRowContext(ctx, q, businessID)
	err := row.Scan(&res.ID, &res.Name, &res.Slug, &res.LogoURL, &res.Currency, &res.Timezone)
	return res, err
}

func (r *Repository) LookupBySlug(ctx context.Context, slug string) (domain.SlugLookupResult, error) {
	var out domain.SlugLookupResult

	const bizQ = `SELECT id, name, COALESCE(slug, ''), COALESCE(logo_url, ''), COALESCE(currency, 'USD'), COALESCE(timezone, 'UTC'), COALESCE(owner_id, 0)
	              FROM businesses WHERE slug = ? AND is_active = TRUE LIMIT 1`
	var ownerID int64
	row := r.db.QueryRowContext(ctx, bizQ, slug)
	if err := row.Scan(&out.Business.ID, &out.Business.Name, &out.Business.Slug, &out.Business.LogoURL, &out.Business.Currency, &out.Business.Timezone, &ownerID); err != nil {
		return out, err
	}

	const locQ = `SELECT id, name, COALESCE(slug, ''), COALESCE(address, ''), COALESCE(timezone, ''), is_active
	              FROM hr_locations WHERE owner_id = ? AND is_active = TRUE ORDER BY is_default DESC, name ASC`
	rows, err := r.db.QueryContext(ctx, locQ, ownerID)
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
	_, err := r.db.ExecContext(ctx, `UPDATE businesses SET slug = ? WHERE id = ?`, slug, businessID)
	return err
}
