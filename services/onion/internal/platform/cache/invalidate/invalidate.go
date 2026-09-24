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
	"github.com/voxire/onion/internal/platform/logger"
)

const PatternInvalidationChannel = "cache-pattern-invalidation"

type PatternInvalidationConsumer struct {
	Patterns []string `json:"patterns"`
}

func (c PatternInvalidationConsumer) Execute(ctx context.Context) error {
	cache := cacher.GetRedisCache()
	if cache == nil {
		return nil
	}

	deleted := 0
	for _, pattern := range c.Patterns {
		deleted += cache.DeleteByPattern(ctx, pattern)
	}
	if len(c.Patterns) > 0 {
		logger.Ctx(ctx).Info("processed async cache invalidation",
			logger.Any("patterns", c.Patterns),
			logger.Int("deleted", deleted),
		)
	}
	return nil
}

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

//nolint:revive // Keep the exported name aligned with the sibling Invalidate API.
func InvalidateAsync(_ context.Context, t keys.InvalidationType, id string) Result {
	patterns := patternsFor(t, id)
	res := Result{Type: t, ID: id, Patterns: patterns}
	if len(patterns) == 0 {
		return res
	}

	cacher.NewPubSub[PatternInvalidationConsumer]().Publish(PatternInvalidationChannel, PatternInvalidationConsumer{
		Patterns: patterns,
	})
	return res
}

func SubscribeAsyncInvalidations(ctx context.Context) {
	cacher.NewPubSub[PatternInvalidationConsumer]().SubscribeContext(ctx, PatternInvalidationChannel)
}

// patternsFor lists every wildcard pattern the type owns. We keep this
// table tiny on purpose — anything domain-specific should re-export its
// own prefix from internal/platform/cache/keys.
func patternsFor(t keys.InvalidationType, id string) []string {
	switch t {
	case keys.InvalidateTeam, keys.InvalidateConfig:
		// Team membership and configuration are bucketed under a numeric
		// business id. Reject malformed ids here rather than constructing a
		// wildcard that can never match a real key.
		idN, err := strconv.ParseInt(id, 10, 64)
		if err != nil || idN <= 0 {
			return nil
		}
		prefix := "team"
		if t == keys.InvalidateConfig {
			prefix = "config"
		}
		return []string{fmt.Sprintf("%sbusiness:%d:%s*", keys.Namespace, idN, prefix)}
	case keys.InvalidateAll:
		return []string{keys.Namespace + "*"}
	}

	idN, err := strconv.ParseInt(id, 10, 64)
	if err != nil || idN <= 0 {
		return nil
	}

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
	case keys.InvalidateSubscription:
		return keys.SubscriptionPatterns(idN)
	}
	return nil
}
