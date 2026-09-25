package cacher

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/voxire/onion/internal/platform/logger"
)

// SRem removes a member from a Redis set.
func (r *redisCache) SRem(ctx context.Context, key string, val []byte) bool {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	command := r.redisCacheCluster.SRem(ctx, key, val)

	if command.Err() != nil {
		logger.Ctx(ctx).Warn("Error while removing from set",
			logger.String("key", key),
			logger.Error(command.Err()),
		)

		return false
	}

	return true
}

// SMembers retrieves all members from a Redis set.
func (r *redisCache) SMembers(ctx context.Context, key string) [][]byte {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	command := r.redisCacheCluster.SMembers(ctx, key)

	redisResults, err := command.Result()
	if err != nil {
		logger.Ctx(ctx).Warn("Error while fetching set members",
			logger.String("key", key),
			logger.Error(command.Err()),
		)
	}

	results := make([][]byte, len(redisResults))

	for i, v := range redisResults {
		results[i] = []byte(v)
	}

	return results
}

// SCard returns the number of members in a Redis set.
func (r *redisCache) SCard(ctx context.Context, key string) (int64, bool) {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	command := r.redisCacheCluster.SCard(ctx, key)

	if command.Err() != nil {
		logger.Ctx(ctx).Warn("Error while getting count of members of set",
			logger.String("key", key),
			logger.Error(command.Err()),
		)

		return 0, false
	}

	return command.Val(), true
}

// DeleteByPattern SCANs for keys matching the glob `pattern` (e.g.
// "cache:user:42:*") and deletes them in batched UNLINK calls. Use this
// for bulk invalidation; for known keys use Delete() instead.
func (r *redisCache) DeleteByPattern(ctx context.Context, pattern string) int {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	if pattern == "" {
		return 0
	}

	const batch = 200
	deleted := GetInMemoryCache(ctx).DeleteByPattern(ctx, pattern)

	for _, currentPattern := range cachePatternVariants(pattern) {
		iter := r.redisCacheCluster.Scan(ctx, 0, currentPattern, batch).Iterator()
		keys := make([]string, 0, batch)

		flush := func() {
			if len(keys) == 0 {
				return
			}
			if err := r.redisCacheCluster.Del(ctx, keys...).Err(); err != nil {
				logger.Ctx(ctx).Warn("redis pattern delete failed",
					logger.Error(err),
					logger.String("pattern", currentPattern),
				)
			} else {
				deleted += len(keys)
			}
			keys = keys[:0]
		}

		for iter.Next(ctx) {
			keys = append(keys, iter.Val())
			if len(keys) >= batch {
				flush()
			}
		}
		flush()

		if err := iter.Err(); err != nil {
			logger.Ctx(ctx).Warn("redis scan failed during pattern delete",
				logger.Error(err),
				logger.String("pattern", currentPattern),
			)
		}
		if ctx.Err() != nil {
			break
		}
	}
	return deleted
}

// Delete deletes one or multiple keys from Redis.
func (r *redisCache) Delete(ctx context.Context, keys []string) {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	pipe := r.redisCacheCluster.Pipeline()

	for _, key := range keys {
		pipe.Del(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Ctx(ctx).Warn("Error while deleting from redis",
			logger.Error(err),
			logger.String("keys", strings.Join(keys, ",")),
		)
	}
}

// HDelete deletes one or multiple fields from a Redis hash set.
func (r *redisCache) HDelete(ctx context.Context, key string, fields []string) {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	cmd := r.redisCacheCluster.HDel(ctx, key, fields...)

	redisError := cmd.Err()
	if redisError != nil {
		logger.Ctx(ctx).Warn("Error while deleting from redis",
			logger.Error(redisError),
			logger.String("key", key),
			logger.String("fields", strings.Join(fields, ",")),
		)
	}
}

// Increment increments the value associated with a key in Redis and returns the new value.
func (r *redisCache) Increment(ctx context.Context, key string, ttl time.Duration) int64 {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	pipe := r.redisCacheCluster.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)

	result, redisError := pipe.Exec(ctx)
	if redisError != nil {
		logger.Ctx(ctx).Warn("Error while incrementing redis key",
			logger.Error(redisError),
			logger.String("key", key),
		)

		return 0
	}

	const expectedIncrResultCount = 2
	if len(result) != expectedIncrResultCount {
		logger.Ctx(ctx).Warn("Error while parsing redis increment result",
			logger.String("key", key),
			logger.Int("result_length", len(result)),
		)

		return 0
	}

	intResult, ok := result[0].(*redis.IntCmd)

	if !ok {
		logger.Ctx(ctx).Warn("Error while parsing redis increment result",
			logger.String("key", key),
			logger.String("result_type", fmt.Sprintf("%T", result[0])),
		)

		return 0
	}

	return intResult.Val()
}

func (r *redisCache) Expire(ctx context.Context, key string, ttl time.Duration) {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	command := r.redisCacheCluster.Expire(ctx, key, ttl)

	if command.Err() != nil {
		logger.Ctx(ctx).Warn("Error while Expire in redis",
			logger.String("key", key),
			logger.Any("ttl", ttl),
			logger.Error(command.Err()),
		)
	}
}

func (r *redisCache) ExpireAt(ctx context.Context, key string, ttl time.Duration) {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	command := r.redisCacheCluster.ExpireAt(ctx, key, time.Now().Add(ttl))

	if command.Err() != nil {
		logger.Ctx(ctx).Warn("Error while ExpireAt in redis",
			logger.String("key", key),
			logger.Any("ttl", ttl),
			logger.Error(command.Err()),
		)
	}
}
