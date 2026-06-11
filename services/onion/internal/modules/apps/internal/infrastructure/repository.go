package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	sq "github.com/Masterminds/squirrel"

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

func (r *Repository) GetUserBusinessID(ctx context.Context, userID int64) (int64, error) {
	var businessID sql.NullInt64
	err := r.pool.QueryRowContext(ctx, "SELECT business_id FROM users WHERE id = ?", userID).Scan(&businessID)
	if err != nil {
		return 0, err
	}
	if !businessID.Valid {
		return 0, nil
	}
	return businessID.Int64, nil
}

// GetEnabledAddons returns addon IDs that are enabled for the given business.
func (r *Repository) GetEnabledAddons(ctx context.Context, businessID int64) map[string]bool {
	// cross-schema read from business_addons — addon enablement is owned by another module
	query, args, err := sq.
		Select("addon_id").
		From("business_addons").
		Where(sq.Eq{"business_id": businessID}).
		Where(sq.Eq{"is_enabled": 1}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("failed to build enabled addons query",
			logger.Any("business_id", businessID),
			logger.Error(err),
		)
		return map[string]bool{}
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
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
	// cross-schema read from user_app_permissions — app permissions are owned by another module
	query, args, err := sq.
		Select("apps").
		From("user_app_permissions").
		Where(sq.Or{
			sq.Eq{"business_id": businessID},
			sq.And{
				sq.Eq{"business_id": nil},
				sq.Expr("owner_id = (SELECT owner_id FROM businesses WHERE id = ? LIMIT 1)", businessID),
			},
		}).
		Where(sq.Eq{"user_id": userID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("failed to build user allowed apps query",
			logger.Any("business_id", businessID),
			logger.Any("user_id", userID),
			logger.Error(err),
		)
		return []string{}
	}

	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&appsJSON)
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
