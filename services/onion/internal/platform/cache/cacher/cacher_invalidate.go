package cacher

import (
	"context"
	"encoding/json"

	"github.com/voxire/onion/internal/platform/logger"
)

// Invalidate invalidates keys from cache.
func (c Cache[M]) Invalidate(ctx context.Context, ids []string) {
	c.Delete(ctx, ids)

	cacheKeys := c.transformListKeys(ids)

	NewPubSub[LocalCacheInvalidationConsumer]().Publish(LocalCacheInvalidationChannel,
		LocalCacheInvalidationConsumer{
			CacheKeys: cacheKeys,
		})
}

// InvalidateScope removes every local and remote cache entry matching each
// supplied scope pattern. Domain services use this for wildcard invalidation
// when a mutation affects a whole user or business cache bucket.
func InvalidateScope(ctx context.Context, patterns ...string) int {
	cache := GetRedisCache()
	if cache == nil {
		return 0
	}

	deleted := 0
	for _, pattern := range patterns {
		deleted += cache.DeleteByPattern(ctx, pattern)
	}
	return deleted
}

// InvalidateTypedScope removes entries matching patterns from a typed cache.
// Typed cache keys include the model identifier, so callers with a legacy or
// model-specific key family must use this variant instead of InvalidateScope.
func InvalidateTypedScope[M any](ctx context.Context, patterns ...string) int {
	cache := GetRedisCache()
	if cache == nil {
		return 0
	}

	deleted := 0
	for _, pattern := range Of[M]().transformListKeys(patterns) {
		deleted += cache.DeleteByPattern(ctx, pattern)
	}
	return deleted
}

func (c Cache[M]) DeleteByPattern(ctx context.Context, patterns []string) int {
	if len(patterns) == 0 {
		return 0
	}

	deleted := 0
	for _, pattern := range c.transformListKeys(patterns) {
		deleted += getRedisCache().DeleteByPattern(ctx, pattern)
	}
	return deleted
}

// SortedSetAdd adds a member with a given score to a Redis sorted set associated with the specified identifier.
func (c Cache[M]) SortedSetAdd(ctx context.Context, id string, val M, score float64, clientOptions ...ClientOption) {
	configuration := getConfiguration(clientOptions...)

	key := c.appendSortedSetIdentifierPrefix(id)

	bytes, err := json.Marshal(val)
	if err != nil {
		logger.Ctx(ctx).Warn("Error while marshalling data to put in redis cache",
			logger.Error(err),
		)

		return
	}

	getRedisCache().ZAdd(ctx, key, bytes, score, configuration.expiry)
}
