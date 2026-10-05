package cacher

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
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
