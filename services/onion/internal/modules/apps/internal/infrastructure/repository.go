package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/voxire/onion/internal/modules/apps/internal/infrastructure/persistence/sqlc"
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

// GetUserAllowedApps returns the list of app IDs a user has permission to access.
// A missing permission row is an empty (and cacheable) result; a query or decode
// failure is returned as an error so it is not cached as an empty catalogue.
func (r *Repository) GetUserAllowedApps(ctx context.Context, businessID, userID int64) ([]string, error) {
	// cross-schema read from user_app_permissions — app permissions are owned by another module
	appsJSON, err := r.queries.GetUserAllowedAppsJSON(ctx, sqlc.GetUserAllowedAppsJSONParams{
		UserID:     userID,
		BusinessID: sql.NullInt64{Int64: businessID, Valid: true},
		ID:         businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []string{}, nil
		}
		logger.Ctx(ctx).Error("failed to query user allowed apps",
			logger.Any("business_id", businessID),
			logger.Any("user_id", userID),
			logger.Error(err),
		)
		return []string{}, err
	}

	var apps []string
	if jsonErr := json.Unmarshal(appsJSON, &apps); jsonErr != nil {
		logger.Ctx(ctx).Warn("failed to unmarshal user allowed apps json",
			logger.Any("business_id", businessID),
			logger.Any("user_id", userID),
			logger.Error(jsonErr),
		)
		return []string{}, jsonErr
	}
	return apps, nil
}
