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
	const q = `SELECT id, name, COALESCE(logo_url, ''), COALESCE(currency, 'USD'), COALESCE(timezone, 'UTC')
	           FROM businesses WHERE id = ? AND is_active = TRUE LIMIT 1`

	var res domain.PublicBusiness
	row := r.db.QueryRowContext(ctx, q, businessID)
	err := row.Scan(&res.ID, &res.Name, &res.LogoURL, &res.Currency, &res.Timezone)
	return res, err
}
