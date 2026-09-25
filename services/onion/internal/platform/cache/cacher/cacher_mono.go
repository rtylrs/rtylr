package cacher

import (
	"context"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/maputils"
)

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
