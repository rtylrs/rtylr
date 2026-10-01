package http

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

const subscriptionAccessCacheTTL = time.Minute

type subscriptionAccess struct {
	Status   string `json:"status"`
	Granted  bool   `json:"granted"`
	TrialEnd string `json:"trial_end,omitempty"`
}

// SubscriptionEnforcementMiddleware blocks authenticated tenant requests when
// the owner's subscription is cancelled or expired. Billing endpoints remain
// available so an owner can recover access, and console administrators are
// never tenant-gated. The decision is cached briefly and invalidated whenever
// subscription state changes.
func SubscriptionEnforcementMiddleware(db *sql.DB, accessPolicy *RouteAccessPolicy) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if accessPolicy != nil && accessPolicy.IsAllowedAnonymously(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			authUser, ok := authPkg.GetAuthUser(r.Context())
			if !ok || authUser == nil {
				next.ServeHTTP(w, r)
				return
			}
			if strings.HasPrefix(r.URL.Path, "/hub/subscription/") || authPkg.IsConsoleAdmin(authUser) {
				next.ServeHTTP(w, r)
				return
			}

			ownerID, err := tenant.ResolveOwnerID(r.Context(), authUser.BusinessID, authUser.UserID)
			if err != nil || ownerID <= 0 {
				writeSubscriptionError(r.Context(), w, http.StatusServiceUnavailable, "Subscription status is temporarily unavailable")
				return
			}

			access, err := cachedSubscriptionAccess(r.Context(), db, ownerID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					next.ServeHTTP(w, r)
					return
				}
				writeSubscriptionError(r.Context(), w, http.StatusServiceUnavailable, "Subscription status is temporarily unavailable")
				return
			}
			if !subscriptionAccessAllowed(access, time.Now().UTC()) {
				writeSubscriptionError(r.Context(), w, http.StatusPaymentRequired, "An active subscription is required")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func cachedSubscriptionAccess(ctx context.Context, db *sql.DB, ownerID int64) (subscriptionAccess, error) {
	if db == nil {
		return subscriptionAccess{}, errors.New("subscription database is unavailable")
	}
	return cacher.Of[subscriptionAccess]().GetOrComputeSingle(
		ctx,
		cacheKeys.SubscriptionAccess(ownerID),
		func(ctx context.Context) (subscriptionAccess, error) {
			var access subscriptionAccess
			var trialEnd sql.NullTime
			err := db.QueryRowContext(dbclient.WithTag(ctx, dbclient.TagMain), `
SELECT s.status,
       EXISTS (
           SELECT 1 FROM manual_access_grants g
           WHERE g.user_id = s.owner_id
             AND g.plan_id = s.plan_id
             AND g.revoked_at IS NULL
             AND (g.expires_at IS NULL OR g.expires_at > CURRENT_TIMESTAMP)
       ),
       s.trial_ends_at
FROM subscriptions s
WHERE s.owner_id = ?`, ownerID).Scan(&access.Status, &access.Granted, &trialEnd)
			if err != nil {
				return subscriptionAccess{}, err
			}
			if trialEnd.Valid {
				access.TrialEnd = trialEnd.Time.UTC().Format(time.RFC3339Nano)
			}
			return access, nil
		},
		cacher.WithExpiry(ctx, subscriptionAccessCacheTTL),
	)
}

func subscriptionAccessAllowed(access subscriptionAccess, now time.Time) bool {
	if access.Granted {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(access.Status)) {
	case "active":
		return true
	case "trial":
		if access.TrialEnd == "" {
			return true
		}
		trialEnd, err := time.Parse(time.RFC3339Nano, access.TrialEnd)
		return err == nil && trialEnd.After(now)
	default:
		return false
	}
}

func writeSubscriptionError(ctx context.Context, w http.ResponseWriter, status int, message string) {
	WriteOneKitError(ctx, w, onekiterrors.New(status, message))
}
