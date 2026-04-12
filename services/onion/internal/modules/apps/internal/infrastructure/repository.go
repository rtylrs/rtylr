package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/voxire/onion/internal/platform/database/dbclient"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		pool: dbclient.Pool(),
	}
}

// GetEnabledAddons returns addon IDs that are enabled for the given business.
func (r *Repository) GetEnabledAddons(ctx context.Context, businessID int64) map[string]bool {
	rows, err := r.pool.QueryContext(ctx,
		"SELECT addon_id FROM business_addons WHERE business_id = ? AND is_enabled = 1",
		businessID,
	)
	if err != nil {
		return map[string]bool{}
	}
	defer rows.Close()

	enabled := make(map[string]bool)
	for rows.Next() {
		var addonID string
		if scanErr := rows.Scan(&addonID); scanErr == nil {
			enabled[addonID] = true
		}
	}
	return enabled
}

// GetUserAllowedApps returns the list of app IDs a user has permission to access.
func (r *Repository) GetUserAllowedApps(ctx context.Context, businessID, userID int64) []string {
	var appsJSON []byte
	err := r.pool.QueryRowContext(ctx,
		"SELECT apps FROM user_app_permissions WHERE (business_id = ? OR owner_id = ?) AND user_id = ?",
		businessID, businessID, userID,
	).Scan(&appsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return []string{}
		}
		return []string{}
	}

	var apps []string
	if jsonErr := json.Unmarshal(appsJSON, &apps); jsonErr != nil {
		return []string{}
	}
	return apps
}
