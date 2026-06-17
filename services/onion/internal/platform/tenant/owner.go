package tenant

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/platform/database/dbclient"
)

// ResolveOwnerID maps an authenticated business context to the canonical tenant
// owner id used by ERP/POS tables. It falls back to the caller user id for older
// tokens, owner accounts without a selected business, and tests without a DB.
func ResolveOwnerID(ctx context.Context, businessID, fallbackUserID int64) int64 {
	if businessID <= 0 {
		return fallbackUserID
	}

	db := dbclient.Pool()
	if db == nil {
		return fallbackUserID
	}

	var ownerID int64
	err := db.QueryRowContext(ctx,
		`SELECT owner_id FROM businesses WHERE id = ? AND is_active = TRUE LIMIT 1`,
		businessID,
	).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fallbackUserID
		}
		return fallbackUserID
	}
	if ownerID <= 0 {
		return fallbackUserID
	}
	return ownerID
}
