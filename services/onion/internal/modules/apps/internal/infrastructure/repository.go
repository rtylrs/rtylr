package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		pool: db,
	}
}

// GetEnabledAddons returns addon IDs that are enabled for the given business.
func (r *Repository) GetEnabledAddons(ctx context.Context, businessID int64) map[string]bool {
	rows, err := r.pool.QueryContext(ctx,
		"SELECT addon_id FROM business_addons WHERE business_id = ? AND is_enabled = 1",
		businessID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("failed to query enabled addons",
			logger.Any("business_id", businessID),
			logger.Error(err),
		)
		return map[string]bool{}
	}
	defer rows.Close()

	enabled := make(map[string]bool)
	for rows.Next() {
		var addonID string
		if scanErr := rows.Scan(&addonID); scanErr != nil {
			logger.Ctx(ctx).Warn("failed to scan enabled addon row",
				logger.Any("business_id", businessID),
				logger.Error(scanErr),
			)
			continue
		}
		enabled[addonID] = true
	}
	if err := rows.Err(); err != nil {
		logger.Ctx(ctx).Error("error iterating enabled addon rows",
			logger.Any("business_id", businessID),
			logger.Error(err),
		)
		return map[string]bool{}
	}
	return enabled
}

// GetUserAllowedApps returns the list of app IDs a user has permission to access.
// Scoped strictly by business_id (the tenant key carried in the JWT) and user_id.
// The previous query also OR-matched owner_id against the same businessID value;
// owner_id stores the owning account's user id, not a business id, so that clause
// could match another tenant's row and leak app permissions across tenants.
func (r *Repository) GetUserAllowedApps(ctx context.Context, businessID, userID int64) []string {
	var appsJSON []byte
	err := r.pool.QueryRowContext(ctx,
		"SELECT apps FROM user_app_permissions WHERE business_id = ? AND user_id = ?",
		businessID, userID,
	).Scan(&appsJSON)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []string{}
		}
		logger.Ctx(ctx).Error("failed to query user allowed apps",
			logger.Any("business_id", businessID),
			logger.Any("user_id", userID),
			logger.Error(err),
		)
		return []string{}
	}

	var apps []string
	if jsonErr := json.Unmarshal(appsJSON, &apps); jsonErr != nil {
		logger.Ctx(ctx).Warn("failed to unmarshal user allowed apps json",
			logger.Any("business_id", businessID),
			logger.Any("user_id", userID),
			logger.Error(jsonErr),
		)
		return []string{}
	}
	return apps
}
