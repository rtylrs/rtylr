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

	if ownerID == 0 {
		return profile, errors.New("ownerID must be non-zero")
	}

	row, err := r.queries.GetBusinessProfileByBusiness(ctx, sqlc.GetBusinessProfileByBusinessParams{OwnerID: ownerID})
	if err == nil {
		profile.ID = row.ID
		profile.Name = row.Name
		profile.SiteName = row.SiteName
		profile.LogoURL = sqlcString(row.LogoUrl)
		profile.ContactEmail = row.ContactEmail
		profile.ContactPhone = row.ContactNumber
		profile.Currency = row.Currency
		profile.Timezone = row.Timezone
		profile.Plan = "free"
		return profile, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("Failed to get business profile", logger.Error(err))
		return profile, err
	}

	cfg, err := r.queries.GetBusinessProfile(ctx, int32(ownerID)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("Failed to get business profile", logger.Error(err))
		return profile, err
	}

	profile.ID = ownerID
	profile.Name = cfg.SiteName
	profile.SiteName = cfg.SiteName
	profile.LogoURL = cfg.LogoUrl
	profile.ContactEmail = cfg.ContactEmail
	profile.ContactPhone = cfg.ContactNumber
	profile.Currency = "USD"
	profile.Timezone = "UTC"
	profile.Plan = "free"

	return profile, nil
}

func (r *Repository) GetPublicBusiness(ctx context.Context, businessID int64) (domain.PublicBusiness, error) {
	// cross-schema read from businesses (shared) — public storefront lookup
	row, err := r.queries.GetPublicBusiness(ctx, businessID)
	if err != nil {
		return domain.PublicBusiness{}, err
	}
	return domain.PublicBusiness{
		ID:       row.ID,
		Name:     row.Name,
		Slug:     row.Slug,
		LogoURL:  row.LogoUrl,
		Currency: row.Currency,
		Timezone: row.Timezone,
	}, nil
}

func (r *Repository) LookupBySlug(ctx context.Context, slug string) (domain.SlugLookupResult, error) {
	var out domain.SlugLookupResult

	// cross-schema read from businesses (shared) — resolve storefront by slug
	biz, err := r.queries.LookupBusinessBySlug(ctx, sql.NullString{String: slug, Valid: slug != ""})
	if err != nil {
		return out, err
	}

	out.Business = domain.PublicBusiness{
		ID:       biz.ID,
		Name:     biz.Name,
		Slug:     biz.Slug,
		LogoURL:  biz.LogoUrl,
		Currency: biz.Currency,
		Timezone: biz.Timezone,
	}

	// cross-schema read from hr_locations (hr_core) — active storefront locations
	rows, err := r.queries.ListPublicLocationsByBusiness(ctx, biz.ID)
	if err != nil {
		return out, err
	}

	for _, row := range rows {
		out.Locations = append(out.Locations, domain.PublicLocation{
			ID:       row.ID,
			Name:     row.Name,
			Slug:     row.Slug,
			Address:  row.Address,
			Timezone: row.Timezone,
			IsActive: row.IsActive,
		})
	}
	return out, nil
}

// UpdateSlug sets the storefront slug of businessID. The row is addressed by
// the caller-supplied owner id, never by the slug itself, so a caller that does
// not own the business updates nothing and reports false.
func (r *Repository) UpdateSlug(ctx context.Context, businessID, ownerUserID int64, slug string) (bool, error) {
	// cross-schema write to businesses (shared) — set storefront slug
	result, err := r.queries.UpdateBusinessSlug(ctx, sqlc.UpdateBusinessSlugParams{
		Slug:    sql.NullString{String: slug, Valid: slug != ""},
		ID:      businessID,
		OwnerID: ownerUserID,
	})
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}

func (r *Repository) RenameBusiness(ctx context.Context, businessID, ownerUserID int64, name string) (bool, error) {
	result, err := r.queries.RenameBusiness(ctx, sqlc.RenameBusinessParams{
		Name:    name,
		ID:      businessID,
		OwnerID: ownerUserID,
	})
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}

func (r *Repository) SetBusinessLogo(ctx context.Context, businessID, ownerUserID int64, logoURL string) (bool, error) {
	result, err := r.queries.SetBusinessLogo(ctx, sqlc.SetBusinessLogoParams{
		LogoUrl: sql.NullString{String: logoURL, Valid: logoURL != ""},
		ID:      businessID,
		OwnerID: ownerUserID,
	})
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}

func (r *Repository) DeactivateBusiness(ctx context.Context, businessID, ownerUserID int64) (bool, error) {
	result, err := r.queries.DeactivateBusiness(ctx, sqlc.DeactivateBusinessParams{
		ID:      businessID,
		OwnerID: ownerUserID,
	})
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}

func (r *Repository) CountActiveOwnedBusinesses(ctx context.Context, ownerUserID int64) (int64, error) {
	return r.queries.CountActiveOwnedBusinesses(ctx, ownerUserID)
}

// ListUserBusinesses returns all businesses the user owns or has an active
// owner/member grant for.
func (r *Repository) ListUserBusinesses(ctx context.Context, userID int64) ([]domain.OwnedBusiness, error) {
	// Cross-schema reads from user_business_access and user_app_permissions
	// provide one switcher for owners and invited team members.
	rows, err := r.queries.ListUserBusinesses(ctx, sqlc.ListUserBusinessesParams{
		OwnerID:   userID,
		OwnerID_2: userID,
		UserID:    userID,
		UserID_2:  userID,
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.OwnedBusiness, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.OwnedBusiness{
			ID:       row.ID,
			Name:     row.Name,
			LogoURL:  row.LogoUrl,
			ParentID: row.ParentID,
			IsActive: row.IsActive.Valid && row.IsActive.Bool,
			IsOwned:  row.IsOwned,
		})
	}
	return result, nil
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

	qtx := r.queries.WithTx(tx)

	res, err := qtx.InsertSubBusiness(ctx, sqlc.InsertSubBusinessParams{
		Name:     name,
		OwnerID:  ownerUserID,
		ParentID: sql.NullInt64{Int64: parentBusinessID, Valid: parentBusinessID != 0},
	})
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
	if err = qtx.GrantUserBusinessAccess(ctx, sqlc.GrantUserBusinessAccessParams{
		UserID:     ownerUserID,
		BusinessID: businessID,
	}); err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	return businessID, tx.Commit()
}

// CheckSubBusinessOwnership validates that parentBusinessID is owned by ownerUserID.
func (r *Repository) CheckSubBusinessOwnership(ctx context.Context, ownerUserID, parentBusinessID int64) (bool, error) {
	cnt, err := r.queries.CheckSubBusinessOwnership(ctx, sqlc.CheckSubBusinessOwnershipParams{
		ID:      parentBusinessID,
		OwnerID: ownerUserID,
	})
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func sqlcString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return ""
}
