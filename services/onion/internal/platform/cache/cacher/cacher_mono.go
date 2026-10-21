package cacher

import (
	"context"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/maputils"
)

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
