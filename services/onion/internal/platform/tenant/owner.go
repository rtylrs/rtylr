package tenant

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
	tenantsqlc "github.com/voxire/onion/internal/platform/tenant/persistence/sqlc"
)

var ErrOwnerLookupFailed = errors.New("failed to resolve business owner")

// ErrNoBusinessScope is returned when no valid business id can be resolved for
// the request. Callers MUST deny the request (e.g. 403) rather than proceed —
// treating a zero/unresolved scope as "no filter" would fail open across
// tenants, and writing rows under a zero scope would corrupt isolation.
var ErrNoBusinessScope = errors.New("no business scope for request")

const businessAccessQuery = `
SELECT 1
FROM businesses b
WHERE b.id = ?
  AND COALESCE(b.is_active, TRUE) = TRUE
  AND (
    b.owner_id = ?
    OR EXISTS (
      SELECT 1
      FROM user_business_access uba
      WHERE uba.user_id = ?
        AND uba.business_id = b.id
    )
    OR EXISTS (
      SELECT 1
      FROM authorization_business_members abm
      WHERE abm.user_id = ?
        AND abm.business_id = b.id
        AND abm.status = 'active'
    )
    OR EXISTS (
      SELECT 1
      FROM user_app_permissions uap
      WHERE uap.user_id = ?
        AND uap.business_id = b.id
        AND uap.is_active = TRUE
    )
  )
LIMIT 1`

func tenantQueries(db *sql.DB) *tenantsqlc.Queries {
	return tenantsqlc.New(db)
}

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

	ownerID, err := tenantQueries(db).GetActiveBusinessOwnerID(ctx, businessID)
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

// ResolveBusinessID maps an authenticated context to the business id used to
// scope tenant data. Unlike ResolveOwnerID (which collapses every business of an
// owner down to one shared owner id and is reserved for billing aggregation),
// this keeps each business isolated.
//
// It accepts the business_id carried on the token only after verifying that the
// authenticated user owns or is granted access to that active business. For
// older tokens or accounts without a selected business it falls back to the
// user's default business (users.business_id), then applies the same check. If
// the database, membership, or business row cannot be verified it returns
// ErrNoBusinessScope — callers must treat that as a denial. It never returns a
// usable id of 0, so tenant data can never be scoped to (or written under) the
// fail-open "business_id = 0" bucket.
func ResolveBusinessID(ctx context.Context, businessID, fallbackUserID int64) (int64, error) {
	return resolveBusinessID(ctx, dbclient.Pool(), businessID, fallbackUserID)
}

func resolveBusinessID(ctx context.Context, db *sql.DB, businessID, fallbackUserID int64) (int64, error) {
	if db == nil || fallbackUserID <= 0 {
		return 0, ErrNoBusinessScope
	}

	resolvedBusinessID := businessID
	if resolvedBusinessID <= 0 {
		resolved, err := tenantQueries(db).GetUserDefaultBusinessID(ctx, fallbackUserID)
		if err != nil {
			logger.Ctx(ctx).Error("failed to resolve default business for user",
				logger.Any("user_id", fallbackUserID),
				logger.Error(err),
			)
			return 0, ErrNoBusinessScope
		}
		if !resolved.Valid || resolved.Int64 <= 0 {
			return 0, ErrNoBusinessScope
		}
		resolvedBusinessID = resolved.Int64
	}

	var authorized int
	err := db.QueryRowContext(
		dbclient.WithTag(ctx, dbclient.TagMain),
		businessAccessQuery,
		resolvedBusinessID,
		fallbackUserID,
		fallbackUserID,
		fallbackUserID,
		fallbackUserID,
	).Scan(&authorized)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(ctx).Error("failed to verify business access",
				logger.Any("business_id", resolvedBusinessID),
				logger.Any("user_id", fallbackUserID),
				logger.Error(err),
			)
		}
		return 0, ErrNoBusinessScope
	}

	return resolvedBusinessID, nil
}

// ResolveOwnedConfigBusinessID returns the business id an owner may read config
// for. When requestedBusinessID is set it must belong to ownerID; otherwise the
// owner's earliest active business is used.
func ResolveOwnedConfigBusinessID(ctx context.Context, ownerID, requestedBusinessID int64) (int64, error) {
	db := dbclient.Pool()
	if db == nil {
		return 0, ErrNoBusinessScope
	}

	queries := tenantQueries(db)
	if requestedBusinessID > 0 {
		resolved, err := queries.GetOwnedActiveBusinessByID(ctx, tenantsqlc.GetOwnedActiveBusinessByIDParams{
			ID:      requestedBusinessID,
			OwnerID: ownerID,
		})
		if err != nil {
			return 0, err
		}
		return resolved, nil
	}

	resolved, err := queries.GetFirstOwnedActiveBusiness(ctx, ownerID)
	if err != nil {
		return 0, err
	}
	return resolved, nil
}
