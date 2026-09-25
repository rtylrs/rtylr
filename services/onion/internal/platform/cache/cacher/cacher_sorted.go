package cacher

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/maputils"
	"github.com/voxire/onion/internal/platform/utils/sliceutils"
)

// SortedSetRemove removes a member from a Redis sorted set associated with the specified identifier.
func (c Cache[M]) SortedSetRemove(ctx context.Context, id string, val M) {
	key := c.appendSortedSetIdentifierPrefix(id)

	bytes, err := json.Marshal(val)
	if err != nil {
		logger.Ctx(ctx).Warn("Error while marshalling data to put in redis cache",
			logger.Error(err),
		)

		return
	}

	getRedisCache().ZRem(ctx, key, bytes)
}

// SortedSetRemoveRangeUpToScore removes a range from a Redis sorted set by score.
func (c Cache[M]) SortedSetRemoveRangeUpToScore(ctx context.Context, id string, upToScore float64) {
	key := c.appendSortedSetIdentifierPrefix(id)
	getRedisCache().ZRemRangeUpToScore(ctx, key, upToScore)
}

// SortedSetRemoveRangeByRank removes a range from a Redis sorted set by rank (inclusive on start and end).
func (c Cache[M]) SortedSetRemoveRangeByRank(ctx context.Context, id string, start, end int64) {
	key := c.appendSortedSetIdentifierPrefix(id)
	getRedisCache().ZRemRangeByRank(ctx, key, start, end)
}

// SortedSetGetAll retrieves all members from a Redis sorted set associated with the specified identifier.
func (c Cache[M]) SortedSetGetAll(ctx context.Context, id string) []M {
	key := c.appendSortedSetIdentifierPrefix(id)
	redisResults := getRedisCache().ZRange(ctx, key, 0, -1)
	results := make([]M, len(redisResults))

	for i, v := range redisResults {
		model := new(M)
		if err := json.Unmarshal(v, model); err != nil {
			logger.Ctx(ctx).Warn("Error while unmarshalling hash data from redis",
				logger.String("key", key),
				logger.Error(err),
			)

			continue
		}

		results[i] = *model
	}

	return results
}

// SortedSetCount returns the number of members in a Redis sorted set associated with the specified identifier.
func (c Cache[M]) SortedSetCount(ctx context.Context, id string) (int64, bool) {
	key := c.appendSortedSetIdentifierPrefix(id)

	return getRedisCache().ZCard(ctx, key)
}

// SetAdd adds a member to a Redis set associated with the specified identifier.
func (c Cache[M]) SetAdd(ctx context.Context, id string, val M, options ...ClientOption) {
	configuration := getConfiguration(options...)

	key := c.appendSetIdentifierPrefix(id)

	bytes, err := json.Marshal(val)
	if err != nil {
		logger.Ctx(ctx).Warn("Error while marshalling data to add to redis set",
			logger.Error(err),
		)

		return
	}

	getRedisCache().SAdd(ctx, key, bytes, configuration.expiry)
}

// SetRemove removes a member from a Redis set associated with the specified identifier.
func (c Cache[M]) SetRemove(ctx context.Context, id string, val M) {
	key := c.appendSetIdentifierPrefix(id)

	bytes, err := json.Marshal(val)
	if err != nil {
		logger.Ctx(ctx).Warn("Error while marshalling data to remove from redis set",
			logger.Error(err),
		)

		return
	}

	getRedisCache().SRem(ctx, key, bytes)
}

// SetGetAll retrieves all members from a Redis set associated with the specified identifier.
func (c Cache[M]) SetGetAll(ctx context.Context, id string) []M {
	key := c.appendSetIdentifierPrefix(id)
	redisResults := getRedisCache().SMembers(ctx, key)
	results := make([]M, len(redisResults))

	for i, v := range redisResults {
		model := new(M)
		if err := json.Unmarshal(v, model); err != nil {
			logger.Ctx(ctx).Warn("Error while unmarshalling data from redis set",
				logger.String("key", key),
				logger.Error(err),
			)

			continue
		}

		results[i] = *model
	}

	return results
}

// SetGetCount retrieves the number of members of a Redis set associated with the specified identifier.
func (c Cache[M]) SetGetCount(ctx context.Context, id string) (int64, bool) {
	key := c.appendSetIdentifierPrefix(id)

	return getRedisCache().SCard(ctx, key)
}

// HashSetIfNotExists sets a value only if the key does not exist.
// if the key exists in the hash, the new value will not be set.
func (c Cache[M]) HashSetIfNotExists(ctx context.Context, id, field string, val M, options ...ClientOption) {
	configuration := getConfiguration(options...)
	key := c.appendHashIdentifierPrefix(id)

	bytes, err := json.Marshal(val)
	if err != nil {
		logger.Ctx(ctx).Warn("Error while marshalling data to add to redis hash set",
			logger.Error(err),
		)

		return
	}

	getRedisCache().HSetNx(ctx, key, field, bytes, configuration.expiry)
}

// Delete deletes one or multiple keys.
func (c Cache[M]) Delete(ctx context.Context, ids []string, options ...ClientOption) {
	configuration := getConfiguration(options...)

	cacheKeys := c.transformListKeys(ids)
	if !configuration.remoteOnly {
		GetInMemoryCache(ctx).Delete(ctx, cacheKeys)
	}

	if !configuration.memoryOnly {
		getRedisCache().Delete(ctx, cacheKeys)
	}
}

func (c Cache[M]) transformListKeys(keys []string) []string {
	return sliceutils.Map(keys, c.appendIdentifierPrefix)
}

func (c Cache[M]) transformMapKeys(keyVals map[string]M) map[string]M {
	return maputils.MapKeys(keyVals, c.appendIdentifierPrefix)
}

var typeNameCache sync.Map

func (c Cache[M]) getModelIdentifier() string {
	if c.namespace != "" {
		return c.namespace
	}
	t := reflect.TypeOf(*new(M))
	if val, ok := typeNameCache.Load(t); ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	name := t.String()
	typeNameCache.Store(t, name)
	return name
}

func (c Cache[M]) appendIdentifierPrefix(s string) string {
	return c.getModelIdentifier() + "-" + s
}

func (c Cache[M]) appendHashIdentifierPrefix(s string) string {
	return c.getModelIdentifier() + "-set-" + s
}

// review.
func (c Cache[M]) appendSetIdentifierPrefix(s string) string {
	return c.getModelIdentifier() + "-redis-set-" + s
}

func (c Cache[M]) appendSortedSetIdentifierPrefix(s string) string {
	return c.getModelIdentifier() + "-sorted_set-" + s
}

func (c Cache[M]) recoverListKeys(keys []string) []string {
	return sliceutils.Map(keys, c.deleteIdentifierPrefix)
}

func (c Cache[M]) recoverMapKeys(keyVals map[string]M) map[string]M {
	return maputils.MapKeys(keyVals, c.deleteIdentifierPrefix)
}

func (c Cache[M]) deleteIdentifierPrefix(s string) string {
	// Strip only the leading type marker; a global replace would also mangle
	// legitimate ids that happen to contain "<type>-" (e.g. Cache[string]
	// holding the id "string-foo").
	return strings.TrimPrefix(s, c.getModelIdentifier()+"-")
}

func cachePatternVariants(pattern string) []string {
	if pattern == "" {
		return nil
	}
	patterns := []string{pattern}
	if strings.HasPrefix(pattern, "cache:") {
		patterns = append(patterns, "*"+pattern)
	}
	return patterns
}

func shouldForceOriginFetch(probability float32) bool {
	if probability <= 0 {
		return false
	}

	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return false
	}

	randomFloat := float32(binary.BigEndian.Uint32(buf[:])) / float32(^uint32(0))
	return randomFloat <= probability
}

// CounterIncrement increments a key in redis and returns its new value.
func CounterIncrement(ctx context.Context, key string, options ...ClientOption) int64 {
	configuration := getConfiguration(options...)

	return getRedisCache().Increment(ctx, key, configuration.expiry)
}

// CounterGet gets a redis counter integer value.
func CounterGet(ctx context.Context, key string) (int, bool) {
	bytes := getRedisCache().Get(ctx, key)
	if bytes == nil {
		return 0, false
	}

	result, err := strconv.Atoi(string(bytes))
	if err != nil {
		logger.Ctx(ctx).Warn("Error while converting incremental cache value to int",
			logger.String("cache_key", key),
			logger.Error(err),
		)

		return 0, false
	}

	return result, true
}

// SetIfNotExists sets a key-value only if the key does not exist.
// if the key exists, the new value will not be set.
// returns false if key already exists.
func (c Cache[M]) SetIfNotExists(ctx context.Context, id string, val M, options ...ClientOption) bool {
	configuration := getConfiguration(options...)
	key := c.appendIdentifierPrefix(id)

	bytes, err := json.Marshal(val)
	if err != nil {
		logger.Ctx(ctx).Warn("Error while marshalling data to add to redis set",
			logger.Error(err),
		)

		return false
	}

	return getRedisCache().SetNx(ctx, key, bytes, configuration.expiry)
}

// Expire sets an expiry for a redis key.
func (c Cache[M]) Expire(ctx context.Context, id string, ttl time.Duration) {
	if id == "" {
		return
	}

	key := c.appendIdentifierPrefix(id)

	getRedisCache().Expire(ctx, key, ttl)
}

// HashExpire sets an expiry for a redis hash set.
func (c Cache[M]) HashExpire(ctx context.Context, id string, ttl time.Duration) {
	if id == "" {
		return
	}

	key := c.appendHashIdentifierPrefix(id)

	getRedisCache().Expire(ctx, key, ttl)
}
