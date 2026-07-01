package tenant

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

var ErrOwnerLookupFailed = errors.New("failed to resolve business owner")

// ResolveOwnerID maps a business id to the account owner user id on that
// business. Prefer ResolveBusinessID for tenant data scoping — owner id is only
// needed for billing aggregation and legacy owner_id columns.
//
// Returns fallbackUserID when businessID is unset, the DB is unavailable, or the
// business row is missing. Returns an error only when the lookup fails for
// reasons other than a missing row.
func ResolveOwnerID(ctx context.Context, businessID, fallbackUserID int64) (int64, error) {
	if businessID <= 0 {
		return fallbackUserID, nil
	}

	db := dbclient.Pool()
	if db == nil {
		return fallbackUserID, nil
	}

	var ownerID int64
	err := db.QueryRowContext(ctx,
		`SELECT owner_id FROM businesses WHERE id = ? AND is_active = TRUE LIMIT 1`,
		businessID,
	).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fallbackUserID, nil
		}
		logger.Ctx(ctx).Error("failed to resolve owner for business",
			logger.Any("business_id", businessID),
			logger.Error(err),
		)
		return 0, ErrOwnerLookupFailed
	}
	if ownerID <= 0 {
		return fallbackUserID, nil
	}
	return ownerID, nil
}
