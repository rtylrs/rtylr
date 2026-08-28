package cacher

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/maputils"
	"github.com/voxire/onion/internal/platform/utils/sliceutils"
)

const (
	CacheTierLocal = "in_memory"
	CacheTierRedis = "redis"
)

const (
	LocalCacheInvalidationChannel = "invalidate-local-cache"
)

const defaultLocalCacheTTL = 10 * time.Minute

// Cache is a generic cache implementation with support for in-memory and Redis storage.
type Cache[M any] struct {
	namespace string
}

// Of returns a Cache instance for the specified type.
func Of[C any]() Cache[C] {
	return Cache[C]{}
}

// OfWithNamespace returns a Cache instance with an explicit namespace prefix.
func OfWithNamespace[C any](namespace string) Cache[C] {
	return Cache[C]{namespace: namespace}
}

func (c Cache[M]) marshal(val M) ([]byte, error) {
	switch scalar := any(val).(type) {
	case int64:
		return []byte(strconv.FormatInt(scalar, 10)), nil
	case string:
		return []byte(scalar), nil
	case bool:
		return []byte(strconv.FormatBool(scalar)), nil
	default:
		return json.Marshal(val)
	}
}

func (c Cache[M]) unmarshal(bytes []byte, model *M) error {
	switch any(*model).(type) {
	case int64:
		value, err := strconv.ParseInt(strings.TrimSpace(string(bytes)), 10, 64)
		if err != nil {
			return err
		}
		return setDecoded(model, value)
	case string:
		// Accept the old JSON representation so rolling deployments can read
		// values written before the scalar fast path was enabled.
		if len(bytes) > 0 && bytes[0] == '"' {
			return json.Unmarshal(bytes, model)
		}
		return setDecoded(model, string(bytes))
	case bool:
		value, err := strconv.ParseBool(strings.TrimSpace(string(bytes)))
		if err != nil {
			return err
		}
		return setDecoded(model, value)
	default:
		return json.Unmarshal(bytes, model)
	}
}

func setDecoded[M any](model *M, value any) error {
	decoded, ok := value.(M)
	if !ok {
		return fmt.Errorf("cache scalar type mismatch: %T", value)
	}
	*model = decoded
	return nil
}

// GetOrComputeLegacy retrieves values for the specified IDs from the cache.
// If a value is not found in the cache, it invokes the provided fetch function to compute the missing values.
//
// Deprecated: use GetOrCompute instead, which supports error handling in fetch function.
func (c Cache[M]) GetOrComputeLegacy(ctx context.Context, ids []string, fetch func(ctx context.Context, ids []string) map[string]M, options ...ClientOption) map[string]M {
	// Wrapper function that satisfies the new 'fetch' signature.
	// It calls the original 'fetch' and returns nil for the error.
	wrappedFetch := func(ctx context.Context, ids []string) (map[string]M, error) {
		result := fetch(ctx, ids)

		return result, nil
	}

	// Delegate all the core work to the new function.
	return c.GetOrCompute(ctx, ids, wrappedFetch, options...)
}

func (c Cache[M]) GetOrCompute(ctx context.Context, ids []string, fetch func(ctx context.Context, ids []string) (map[string]M, error), options ...ClientOption) map[string]M {
	result, err := c.GetOrComputeWithError(ctx, ids, fetch, options...)
	if err != nil {
		logger.Ctx(ctx).Error("Error in GetOrCompute, returning partial results", logger.Error(err))
	}

	return result
}

// GetOrComputeKeys is a more flexible version that takes specific cache keys and returns values for those keys.
func (c Cache[M]) GetOrComputeKeys(ctx context.Context, keys []string, fetch func(ctx context.Context, missingKeys []string) (map[string]M, error), options ...ClientOption) (map[string]M, error) {
	configuration := getConfiguration(options...)
	forceOriginFetch := shouldForceOriginFetch(configuration.evictionProbability)

	var (
		remaining    map[string]M
		results      map[string]M
		keysNotFound []string
		err          error
	)

	keysNotFound = keys
	if !forceOriginFetch {
		// Transform keys to cache keys for lookup
		cacheKeys := c.transformListKeys(keys)
		cacheResults, cacheKeysNotFound := c.getMultiByTransformedKeys(ctx, cacheKeys, options...)
		results = c.recoverMapKeys(cacheResults)
		keysNotFound = c.recoverListKeys(cacheKeysNotFound)
	}

	if len(keysNotFound) > 0 {
		ctxWithoutCancel := context.WithoutCancel(ctx)

		remaining, err = fetch(ctxWithoutCancel, keysNotFound)
		if err != nil {
			// Don't log here — the cache can't tell an expected
			// sentinel (sql.ErrNoRows, domain.ErrXNotFound) from a
			// real failure. Callers already receive the error and
			// decide whether to alert, swallow, or translate to a
			// 4xx. Logging at ERROR here promoted normal "not found"
			// hits into remote error metrics.
			return results, err
		}

		c.PutMulti(ctxWithoutCancel, remaining, options...)
		results = maputils.Merge(results, remaining)
	}

	return results, nil
}

func (c Cache[M]) GetOrComputeWithError(ctx context.Context, ids []string, fetch func(ctx context.Context, ids []string) (map[string]M, error), options ...ClientOption) (map[string]M, error) {
	configuration := getConfiguration(options...)
	forceOriginFetch := shouldForceOriginFetch(configuration.evictionProbability)

	var (
		remaining    map[string]M
		results      map[string]M
		keysNotFound []string
		err          error
	)

	keysNotFound = ids
	if !forceOriginFetch {
		results, keysNotFound = c.GetMulti(ctx, ids, options...)
	}

	if len(keysNotFound) > 0 {
		ctxWithoutCancel := context.WithoutCancel(ctx)

		remaining, err = fetch(ctxWithoutCancel, keysNotFound)
		if err != nil {
			// Don't log here — the cache can't tell an expected
			// sentinel (sql.ErrNoRows, domain.ErrXNotFound) from a
			// real failure. Callers already receive the error and
			// decide whether to alert, swallow, or translate to a
			// 4xx. Logging at ERROR here promoted normal "not found"
			// hits into remote error metrics.
			return results, err
		}

		c.PutMulti(ctxWithoutCancel, remaining, options...)
		results = maputils.Merge(results, remaining)
	}

	return results, nil
}

// GetOrComputeSingle is a convenience method for single key operations with error handling.
func (c Cache[M]) GetOrComputeSingle(ctx context.Context, key string, fetch func(ctx context.Context) (M, error), options ...ClientOption) (M, error) {
	configuration := getConfiguration(options...)
	if !shouldForceOriginFetch(configuration.evictionProbability) {
		if value, found := c.Get(ctx, key, options...); found {
			return value, nil
		}
	}

	ctxWithoutCancel := context.WithoutCancel(ctx)
	value, err := fetch(ctxWithoutCancel)
	if err != nil {
		return *new(M), err
	}

	c.Put(ctxWithoutCancel, key, value, options...)
	return value, nil
}

// MonoGetOrCompute is a convenience method for calling GetOrComputeLegacy on a single key.
func (c Cache[M]) MonoGetOrCompute(ctx context.Context, key string, fetch func(ctx context.Context, key string) (M, bool), options ...ClientOption) (M, bool) {
	ids := []string{key}

	fetchFunc := func(ctx context.Context, _ []string) map[string]M {
		val, ok := fetch(ctx, key)

		if !ok {
			return map[string]M{}
		}

		return map[string]M{key: val}
	}

	results := c.GetOrComputeLegacy(ctx, ids, fetchFunc, options...)
	val, ok := results[key]

	return val, ok
}

// Put sets or updates a key-value pair in cache.
func (c Cache[M]) Put(ctx context.Context, key string, val M, options ...ClientOption) {
	c.PutMulti(ctx, map[string]M{key: val}, options...)
}

// PutMulti sets or updates multiple key-value pairs in cache.
func (c Cache[M]) PutMulti(ctx context.Context, keyVals map[string]M, options ...ClientOption) {
	configuration := getConfiguration(options...)

	if len(keyVals) == 0 {
		return
	}

	cacheKeyVals := c.transformMapKeys(keyVals)

	marshalledKeyVals := maputils.Map(cacheKeyVals, func(val M) []byte {
		bytes, err := c.marshal(val)
		if err != nil {
			logger.Ctx(ctx).Warn("Error while marshalling data to put in cache",
				logger.Error(err),
			)

			return nil
		}

		return bytes
	})

	shouldMemory := !configuration.remoteOnly
	shouldRemote := !configuration.memoryOnly

	if shouldRemote {
		getRedisCache().PutMulti(ctx, marshalledKeyVals, configuration.expiry)
	}

	if shouldMemory {
		GetInMemoryCache(ctx).PutMulti(ctx, marshalledKeyVals, configuration.expiry)
	}
}

// Get retrieves a key-value pair from cache.
func (c Cache[M]) Get(ctx context.Context, key string, options ...ClientOption) (M, bool) {
	configuration := getConfiguration(options...)
	cacheKey := c.appendIdentifierPrefix(key)

	var bytes []byte
	if !configuration.remoteOnly {
		bytes = GetInMemoryCache(ctx).Get(ctx, cacheKey)
	}

	if bytes == nil && !configuration.memoryOnly {
		bytes = getRedisCache().Get(ctx, cacheKey)
		if bytes != nil && !configuration.remoteOnly {
			GetInMemoryCache(ctx).Put(ctx, cacheKey, bytes, defaultLocalCacheTTL)
		}
	}

	if bytes == nil {
		return *new(M), false
	}

	model := new(M)
	if err := c.unmarshal(bytes, model); err != nil {
		logger.Ctx(ctx).Warn("Error while unmarshalling data from cache store",
			logger.String("key", cacheKey),
			logger.Error(err),
		)

		return *new(M), false
	}

	return *model, true
}

// GetAndDelete atomically consumes a cached value when Redis is available and
// removes the local replica as well. It is intended for one-time credentials.
func (c Cache[M]) GetAndDelete(ctx context.Context, key string, options ...ClientOption) (M, bool) {
	configuration := getConfiguration(options...)
	cacheKey := c.appendIdentifierPrefix(key)
	var bytes []byte

	if !configuration.memoryOnly {
		bytes = getRedisCache().GetDel(ctx, cacheKey)
		if bytes != nil && !configuration.remoteOnly {
			GetInMemoryCache(ctx).Delete(ctx, []string{cacheKey})
		}
	}
	if bytes == nil && !configuration.remoteOnly {
		bytes = GetInMemoryCache(ctx).GetDel(ctx, cacheKey)
	}
	if bytes == nil {
		return *new(M), false
	}

	model := new(M)
	if err := c.unmarshal(bytes, model); err != nil {
		logger.Ctx(ctx).Warn("Error while unmarshalling consumed cache value",
			logger.String("key", cacheKey),
			logger.Error(err),
		)
		return *new(M), false
	}
	return *model, true
}

// GetMulti retrieves multiple key-value pairs from cache.
func (c Cache[M]) getMultiByTransformedKeys(ctx context.Context, transformedKeys []string, options ...ClientOption) (results map[string]M, notFound []string) {
	configuration := getConfiguration(options...)

	results = make(map[string]M, len(transformedKeys))
	if len(transformedKeys) == 0 {
		return results, nil
	}

	fromMemory := make(map[string][]byte)
	fromRedis := make(map[string][]byte)
	remainingKeys := transformedKeys

	if !configuration.remoteOnly {
		fromMemory, remainingKeys = GetInMemoryCache(ctx).GetMulti(ctx, transformedKeys)
	}

	if !configuration.memoryOnly {
		var remainingKeysFromRedis []string
		fromRedis, remainingKeysFromRedis = getRedisCache().GetMulti(ctx, remainingKeys)
		remainingKeys = remainingKeysFromRedis
	}

	if !configuration.remoteOnly && !configuration.memoryOnly {
		GetInMemoryCache(ctx).PutMulti(ctx, fromRedis, defaultLocalCacheTTL)
	}

	for _, key := range transformedKeys {
		bytes := fromMemory[key]
		if bytes == nil {
			bytes = fromRedis[key]
		}

		if bytes == nil {
			continue
		}

		model := new(M)
		if err := c.unmarshal(bytes, model); err != nil {
			logger.Ctx(ctx).Warn("Error while unmarshalling data from cache store",
				logger.String("key", key),
				logger.Error(err),
			)

			continue
		}

		results[key] = *model
	}

	return results, remainingKeys
}

func (c Cache[M]) GetMulti(ctx context.Context, keys []string, options ...ClientOption) (results map[string]M, notFound []string) {
	configuration := getConfiguration(options...)

	cacheKeys := c.transformListKeys(keys)

	results = make(map[string]M, len(cacheKeys))
	if len(cacheKeys) == 0 {
		return results, nil
	}

	fromMemory := make(map[string][]byte)
	fromRedis := make(map[string][]byte)
	remainingKeys := cacheKeys

	if !configuration.remoteOnly {
		fromMemory, remainingKeys = GetInMemoryCache(ctx).GetMulti(ctx, cacheKeys)
	}

	if !configuration.memoryOnly {
		var remainingKeysFromRedis []string
		fromRedis, remainingKeysFromRedis = getRedisCache().GetMulti(ctx, remainingKeys)
		remainingKeys = remainingKeysFromRedis
	}

	if !configuration.remoteOnly && !configuration.memoryOnly {
		GetInMemoryCache(ctx).PutMulti(ctx, fromRedis, defaultLocalCacheTTL)
	}

	for _, key := range cacheKeys {
		bytes := fromMemory[key]
		if bytes == nil {
			bytes = fromRedis[key]
		}

		if bytes == nil {
			continue
		}

		model := new(M)
		if err := c.unmarshal(bytes, model); err != nil {
			logger.Ctx(ctx).Warn("Error while unmarshalling data from cache store",
				logger.String("key", key),
				logger.Error(err),
			)

			continue
		}

		results[key] = *model
	}

	results = c.recoverMapKeys(results)
	notFound = c.recoverListKeys(remainingKeys)

	return results, notFound
}

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
