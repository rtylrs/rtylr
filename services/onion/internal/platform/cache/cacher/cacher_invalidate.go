package cacher

import (
	"context"
	"encoding/json"
	"time"

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

// HashPut sets or updates a key-value pair in a Redis hash.
func (c Cache[M]) HashPut(ctx context.Context, id, hashKey string, val M, options ...ClientOption) {
	configuration := getConfiguration(options...)

	if id == "" || hashKey == "" {
		return
	}

	key := c.appendHashIdentifierPrefix(id)

	bytes, err := json.Marshal(val)
	if err != nil {
		logger.Ctx(ctx).Warn("Error while marshalling data to put in redis cache",
			logger.Error(err),
		)

		return
	}

	getRedisCache().HSet(ctx, key, hashKey, bytes, configuration.expiry)
}

// HashFieldExpire sets an expiry for a field in a redis hash set.
func (c Cache[M]) HashFieldExpire(ctx context.Context, id, field string, ttl time.Duration) {
	if id == "" || field == "" {
		return
	}

	key := c.appendHashIdentifierPrefix(id)

	getRedisCache().HFieldExpire(ctx, key, field, ttl)
}

// HashPutWithFieldExpiry sets or updates a key-value pair in a Redis hash, in addition to adding an expiry to the field in the hashset.
func (c Cache[M]) HashPutWithFieldExpiry(ctx context.Context, id, hashKey string, val M, fieldExpiry time.Duration, options ...ClientOption) {
	configuration := getConfiguration(options...)

	if id == "" || hashKey == "" {
		return
	}

	key := c.appendHashIdentifierPrefix(id)

	bytes, err := json.Marshal(val)
	if err != nil {
		logger.Ctx(ctx).Warn("Error while marshalling data to put in redis cache",
			logger.Error(err),
		)

		return
	}

	getRedisCache().HSetWithFieldExpiry(ctx, key, hashKey, bytes, fieldExpiry, configuration.expiry)
}

// HashDelete deletes a key in a Redis Hash.
func (c Cache[M]) HashDelete(ctx context.Context, id, hashKey string) {
	if id == "" || hashKey == "" {
		return
	}

	key := c.appendHashIdentifierPrefix(id)

	getRedisCache().HDel(ctx, key, hashKey)
}

// HashGet retrieves a value of a certain field in a Redis hash.
func (c Cache[M]) HashGet(ctx context.Context, id, field string, _ ...ClientOption) (M, bool) {
	key := c.appendHashIdentifierPrefix(id)
	bytes := getRedisCache().HGet(ctx, key, field)

	if bytes == nil {
		return *new(M), false
	}

	model := new(M)
	if err := c.unmarshal(bytes, model); err != nil {
		logger.Ctx(ctx).Warn("Error while unmarshalling data from redis",
			logger.String("key", key),
			logger.String("field", field),
			logger.Error(err),
		)

		return *new(M), false
	}

	return *model, true
}

// HashGetAll retrieves all key-value pairs of a Redis hash.
func (c Cache[M]) HashGetAll(ctx context.Context, id string, _ ...ClientOption) map[string]M {
	key := c.appendHashIdentifierPrefix(id)
	redisResults := getRedisCache().HGetAll(ctx, key)
	results := make(map[string]M, len(redisResults))

	for k, v := range redisResults {
		model := new(M)
		if err := c.unmarshal(v, model); err != nil {
			logger.Ctx(ctx).Warn("Error while unmarshalling hash data from redis",
				logger.String("key", key),
				logger.Error(err),
			)
		} else {
			results[k] = *model
		}
	}

	return results
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

// SortedSetAddKeepLastByRank adds a member with a given score to a Redis sorted set associated with the specified identifier,
// while ensuring that only the last k elements are kept in the sorted set by removing the oldest elements.
func (c Cache[M]) SortedSetAddKeepLastByRank(ctx context.Context, id string, val M, score float64, k int64, clientOptions ...ClientOption) {
	c.SortedSetAdd(ctx, id, val, score, clientOptions...)
	// Remove the oldest k elements from the sorted set to keep only the last k elements.
	// -(k+1) is the inclusive backward index in Redis, so it removes the first k elements.
	c.SortedSetRemoveRangeByRank(ctx, id, 0, -(k + 1))
}
