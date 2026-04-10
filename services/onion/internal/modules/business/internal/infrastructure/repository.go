package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/business/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		pool: dbclient.Pool(),
	}
}

func (r *Repository) GetProfile(ctx context.Context, ownerID int64) (domain.BusinessProfile, error) {
	var profile domain.BusinessProfile

	query := `
		SELECT COALESCE(site_name,''), COALESCE(logo_url,''), COALESCE(contact_email,''), COALESCE(contact_number,'')
		FROM configs
		WHERE owner_id = ?
		LIMIT 1`

	var siteName, logoURL, contactEmail, contactPhone string
	err := r.pool.QueryRowContext(ctx, query, ownerID).Scan(&siteName, &logoURL, &contactEmail, &contactPhone)
	if err != nil && err != sql.ErrNoRows {
		logger.Ctx(ctx).Error("Failed to get business profile", logger.Error(err))
		return profile, err
	}

	profile.ID = ownerID
	profile.SiteName = siteName
	profile.LogoURL = logoURL
	profile.ContactEmail = contactEmail
	profile.ContactPhone = contactPhone
	profile.Currency = "USD"
	profile.Timezone = "UTC"
	profile.Plan = "free"

	return profile, nil
}
