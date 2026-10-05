package cacher

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"reflect"
	"strings"
	"sync"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/maputils"
	"github.com/voxire/onion/internal/platform/utils/sliceutils"
)

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

func (c Cache[M]) appendSortedSetIdentifierPrefix(s string) string {
	return c.getModelIdentifier() + "-sorted_set-" + s
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
