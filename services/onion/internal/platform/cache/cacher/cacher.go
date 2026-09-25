package cacher

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/maputils"
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
