// Package invalidate maps high-level invalidation requests
// (type + id) onto the concrete cache-key patterns owned by each domain.
// It's the bridge between the public /invalidate-cache endpoint and the
// individual services.
package invalidate

import (
	"context"
	"fmt"
	"strconv"

	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/cache/keys"
)

// Result reports what an invalidation request affected — useful for the
// endpoint's response body and for log/telemetry breadcrumbs.
type Result struct {
	Type     keys.InvalidationType `json:"type"`
	ID       string                `json:"id"`
	Patterns []string              `json:"patterns"`
	Deleted  int                   `json:"deleted"`
}

// Invalidate wipes every cache entry the named scope touches. Returns the
// patterns it walked + the number of keys it actually deleted.
func Invalidate(ctx context.Context, t keys.InvalidationType, id string) Result {
	patterns := patternsFor(t, id)
	res := Result{Type: t, ID: id, Patterns: patterns}

	cache := cacher.GetRedisCache()
	if cache == nil {
		return res
	}

	for _, p := range patterns {
		res.Deleted += cache.DeleteByPattern(ctx, p)
	}
	return res
}

// patternsFor lists every wildcard pattern the type owns. We keep this
// table tiny on purpose — anything domain-specific should re-export its
// own prefix from internal/platform/cache/keys.
func patternsFor(t keys.InvalidationType, id string) []string {
	idN, _ := strconv.ParseInt(id, 10, 64)
	switch t {
	case keys.InvalidateUser:
		return []string{keys.UserPrefix(idN)}
	case keys.InvalidateBusiness:
		return []string{keys.BusinessPrefix(idN)}
	case keys.InvalidateOrder:
		return []string{keys.OrderPrefix(idN)}
	case keys.InvalidateProduct:
		return []string{keys.ProductPrefix(idN)}
	case keys.InvalidateCustomer:
		return []string{keys.CustomerPrefix(idN)}
	case keys.InvalidateMenu:
		return []string{keys.MenuPrefix(idN)}
	case keys.InvalidateTeam:
		// Team membership is bucketed under the business owner.
		return []string{fmt.Sprintf("%sbusiness:%s:team*", keys.Namespace, id)}
	case keys.InvalidateConfig:
		return []string{fmt.Sprintf("%sbusiness:%s:config*", keys.Namespace, id)}
	case keys.InvalidateSubscription:
		return []string{keys.SubscriptionPrefix(idN)}
	case keys.InvalidateAll:
		return []string{keys.Namespace + "*"}
	}
	return nil
}
