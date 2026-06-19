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

	if err := r.db.QueryRowContext(ctx, `
		SELECT
			b.id,
			b.name,
			COALESCE(NULLIF(c.site_name, ''), b.name) AS site_name,
			COALESCE(NULLIF(c.logo_url, ''), COALESCE(b.logo_url, '')) AS logo_url,
			COALESCE(c.contact_email, '') AS contact_email,
			COALESCE(c.contact_number, '') AS contact_number,
			COALESCE(b.currency, 'USD') AS currency,
			COALESCE(b.timezone, 'UTC') AS timezone
		FROM businesses b
		LEFT JOIN config c ON c.business_id = b.id
		WHERE (b.id = ? OR b.owner_id = ?) AND b.is_active = TRUE
		ORDER BY b.id ASC
		LIMIT 1
	`, ownerID, ownerID).Scan(
		&profile.ID,
		&profile.Name,
		&profile.SiteName,
		&profile.LogoURL,
		&profile.ContactEmail,
		&profile.ContactPhone,
		&profile.Currency,
		&profile.Timezone,
	); err == nil {
		profile.Plan = "free"
		return profile, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("Failed to get business profile", logger.Error(err))
		return profile, err
	}

	row, err := r.queries.GetBusinessProfile(ctx, int32(ownerID))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("Failed to get business profile", logger.Error(err))
		return profile, err
	}

	profile.ID = ownerID
	profile.Name = row.SiteName
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

// ListUserBusinesses returns all businesses the user owns (businesses.owner_id = userID)
// or has explicit access to via user_business_access.
func (r *Repository) ListUserBusinesses(ctx context.Context, userID int64) ([]domain.OwnedBusiness, error) {
	// cross-schema read from user_business_access — multi-business ownership
	q, args, err := sq.Select(
		"b.id", "b.name", "COALESCE(b.logo_url, '')", "COALESCE(b.parent_id, 0)", "b.is_active",
	).
		From("businesses b").
		Where(sq.Or{
			sq.Eq{"b.owner_id": userID},
			sq.Expr("EXISTS (SELECT 1 FROM user_business_access uba WHERE uba.user_id = ? AND uba.business_id = b.id)", userID),
		}).
		Where(sq.Eq{"b.is_active": true}).
		OrderBy("b.id ASC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []domain.OwnedBusiness
	for rows.Next() {
		var b domain.OwnedBusiness
		if err := rows.Scan(&b.ID, &b.Name, &b.LogoURL, &b.ParentID, &b.IsActive); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

// CreateSubBusiness creates a new child business and grants the owner access to it.
func (r *Repository) CreateSubBusiness(ctx context.Context, name string, ownerUserID, parentBusinessID int64) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	res, err := tx.ExecContext(ctx, `INSERT INTO businesses (name, owner_id, parent_id) VALUES (?, ?, ?)`,
		name, ownerUserID, parentBusinessID)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	businessID, err := res.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	// cross-schema write to user_business_access — grant owner access to the new sub-business
	_, err = tx.ExecContext(ctx, `INSERT INTO user_business_access (user_id, business_id, role) VALUES (?, ?, 'owner')`,
		ownerUserID, businessID)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	return businessID, tx.Commit()
}

// CheckSubBusinessOwnership validates that parentBusinessID is owned by ownerUserID.
func (r *Repository) CheckSubBusinessOwnership(ctx context.Context, ownerUserID, parentBusinessID int64) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM businesses WHERE id = ? AND owner_id = ? AND is_active = TRUE`,
		parentBusinessID, ownerUserID).Scan(&count)
	return count > 0, err
}
