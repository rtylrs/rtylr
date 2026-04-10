package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/business/internal/domain"
	"github.com/voxire/onion/internal/modules/business/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		queries: sqlc.New(dbclient.Pool()),
	}
}

func (r *Repository) GetProfile(ctx context.Context, ownerID int64) (domain.BusinessProfile, error) {
	var profile domain.BusinessProfile

	row, err := r.queries.GetBusinessProfile(ctx, int32(ownerID))
	if err != nil && err != sql.ErrNoRows {
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
