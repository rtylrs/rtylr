package tenant

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/platform/database/dbclient"
)

// ErrNoBusinessScope is returned when no valid business id can be resolved for
// the request. Callers MUST deny the request (e.g. 403) rather than proceed —
// treating a zero/unresolved scope as "no filter" would fail open across
// tenants, and writing rows under a zero scope would corrupt isolation.
var ErrNoBusinessScope = errors.New("no business scope for request")

// ResolveBusinessID maps an authenticated context to the business id used to
// scope tenant data. Unlike ResolveOwnerID (which collapses every business of an
// owner down to one shared owner id and is reserved for billing aggregation),
// this keeps each business isolated.
//
// It prefers the business_id carried on the token. For older tokens or owner
// accounts without a selected business it falls back to the user's default
// business (users.business_id). If neither yields a positive business id it
// returns ErrNoBusinessScope — callers must treat that as a denial. It never
// returns a usable id of 0, so tenant data can never be scoped to (or written
// under) the fail-open "business_id = 0" bucket.
func ResolveBusinessID(ctx context.Context, businessID, fallbackUserID int64) (int64, error) {
	if businessID > 0 {
		return businessID, nil
	}

	if fallbackUserID > 0 {
		if db := dbclient.Pool(); db != nil {
			var resolved int64
			if err := db.QueryRowContext(ctx,
				`SELECT business_id FROM users WHERE id = ? LIMIT 1`,
				fallbackUserID,
			).Scan(&resolved); err == nil && resolved > 0 {
				return resolved, nil
			}
		}
	}

	return 0, ErrNoBusinessScope
}
