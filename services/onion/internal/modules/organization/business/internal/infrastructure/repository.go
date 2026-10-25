package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/organization/business/internal/domain"
	"github.com/voxire/onion/internal/modules/organization/business/internal/infrastructure/persistence/sqlc"
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

// GetProfile reads an active business's profile. An unknown or inactive
// business returns sql.ErrNoRows.
func (r *Repository) GetProfile(ctx context.Context, businessID int64) (domain.BusinessProfile, error) {
	if businessID <= 0 {
		return domain.BusinessProfile{}, sql.ErrNoRows
	}
	row, err := r.queries.GetBusinessProfileByBusiness(ctx, businessID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(ctx).Error("Failed to get business profile", logger.Error(err))
		}
		return domain.BusinessProfile{}, err
	}
	return domain.BusinessProfile{
		ID:           row.ID,
		Name:         row.Name,
		SiteName:     row.SiteName,
		LogoURL:      sqlcString(row.LogoUrl),
		ContactEmail: row.ContactEmail,
		ContactPhone: row.ContactNumber,
		Currency:     row.Currency,
		Timezone:     row.Timezone,
		Plan:         "free",
	}, nil
}

func (r *Repository) GetBusinessOwnerID(ctx context.Context, businessID int64) (int64, error) {
	return r.queries.GetBusinessOwnerID(ctx, businessID)
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

// ListUserBusinesses returns all businesses the user owns or has active membership in.
func (r *Repository) ListUserBusinesses(ctx context.Context, userID int64) ([]domain.OwnedBusiness, error) {
	// The normalized membership table is the shared source for business access.
	rows, err := r.queries.ListUserBusinesses(ctx, sqlc.ListUserBusinessesParams{
		UserID: userID,
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

func (r *Repository) ReportingPreferences(ctx context.Context, businessID int64) (currency, timezone string, err error) {
	row, err := r.queries.GetReportingPreferences(ctx, businessID)
	return sqlcString(row.Currency), sqlcString(row.Timezone), err
}
