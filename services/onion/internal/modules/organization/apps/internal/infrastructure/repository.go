package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/organization/apps/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		pool:    db,
	}
}

func (r *Repository) GetUserBusinessID(ctx context.Context, userID int64) (int64, error) {
	businessID, err := tenant.ResolveBusinessID(ctx, 0, userID)
	if err != nil {
		if errors.Is(err, tenant.ErrNoBusinessScope) {
			return 0, nil
		}
		return 0, err
	}
	return businessID, nil
}

// GetEnabledAddons returns addon IDs that are enabled for the given business.
// The error is returned rather than folded into an empty map so the service's
// cache fetch closure can refuse to cache a transient failure.
func (r *Repository) GetEnabledAddons(ctx context.Context, businessID int64) (map[string]bool, error) {
	// cross-schema read from business_addons — addon enablement is owned by another module
	rows, err := r.queries.ListEnabledAddonIDs(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to query enabled addons",
			logger.Any("business_id", businessID),
			logger.Error(err),
		)
		return map[string]bool{}, err
	}

	enabled := make(map[string]bool, len(rows))
	for _, addonID := range rows {
		enabled[addonID] = true
	}
	return enabled, nil
}
