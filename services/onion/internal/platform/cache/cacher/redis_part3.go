package cacher

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/voxire/onion/internal/platform/logger"
)

// HSetNx sets the value of a field in a Redis hash set only if the field does not exist.
func (r *redisCache) HSetNx(ctx context.Context, key, field string, val []byte, ttl time.Duration) {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	pipe := r.redisCacheCluster.Pipeline()
	pipe.HSetNX(ctx, key, field, val)
	pipe.Expire(ctx, key, ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Ctx(ctx).Warn("Error while incrementing redis key",
			logger.Error(err),
			logger.String("key", key),
		)
	}
}

// SetNx sets a key-value pair only if the key does not exist.
// returns false if key already exists.
func (r *redisCache) SetNx(ctx context.Context, key string, val []byte, ttl time.Duration) bool {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	command := r.redisCacheCluster.SetNX(ctx, key, val, ttl)

	didSet, err := command.Result()
	if err != nil {
		logger.Ctx(ctx).Warn("Error while setNX redis key",
			logger.Error(err),
			logger.String("key", key),
		)
	}

	return didSet
}

// ZAdd adds a member with a specified score to a Redis sorted set.
// The score determines the position of the member in the sorted set,
// with lower scores coming before higher scores.
func (r *redisCache) ZAdd(ctx context.Context, key string, val []byte, score float64, ttl time.Duration) bool {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	pipe := r.redisCacheCluster.Pipeline()

	zMember := redis.Z{
		Score:  score,
		Member: val,
	}
	pipe.ZAdd(ctx, key, zMember)
	pipe.Expire(ctx, key, ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Ctx(ctx).Warn("Error while adding to sorted set",
			logger.String("key", key),
			logger.Error(err),
		)

		return false
	}

	return true
}

// ZRem removes a member from a Redis sorted set.
func (r *redisCache) ZRem(ctx context.Context, key string, val []byte) bool {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	command := r.redisCacheCluster.ZRem(ctx, key, val)

	if command.Err() != nil {
		logger.Ctx(ctx).Warn("Error while removing from sorted set",
			logger.String("key", key),
			logger.Error(command.Err()),
		)

		return false
	}

	return true
}

// ZRemRangeByScore removes a range from a Redis sorted set by score.
func (r *redisCache) ZRemRangeByScore(ctx context.Context, key, minVal, maxVal string) bool {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	command := r.redisCacheCluster.ZRemRangeByScore(ctx, key, minVal, maxVal)

	if command.Err() != nil {
		logger.Ctx(ctx).Warn("Error while removing from sorted set",
			logger.String("key", key),
			logger.Error(command.Err()),
		)

		return false
	}

	return true
}

// ZRemRangeUpToScore removes a range from a Redis sorted set up to a score.
func (r *redisCache) ZRemRangeUpToScore(ctx context.Context, key string, upToScore float64) bool {
	return r.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatFloat(upToScore, 'f', -1, 64))
}

// ZRemRangeByRank removes a range from a Redis sorted set by rank.
// start and end are inclusive indices.
func (r *redisCache) ZRemRangeByRank(ctx context.Context, key string, start, end int64) bool {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	command := r.redisCacheCluster.ZRemRangeByRank(ctx, key, start, end)

	if command.Err() != nil {
		logger.Ctx(ctx).Warn("Error while removing from sorted set",
			logger.String("key", key),
			logger.Error(command.Err()),
		)

		return false
	}

	return true
}

// ZCard returns the number of members in a Redis sorted set.
func (r *redisCache) ZCard(ctx context.Context, key string) (int64, bool) {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	command := r.redisCacheCluster.ZCard(ctx, key)

	if command.Err() != nil {
		logger.Ctx(ctx).Warn("Error while getting count of members of sorted set",
			logger.String("key", key),
			logger.Error(command.Err()),
		)

		return 0, false
	}

	return command.Val(), true
}

// ZRange retrieves a range of members from a Redis sorted set.
func (r *redisCache) ZRange(ctx context.Context, key string, start, stop int64) [][]byte {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	command := r.redisCacheCluster.ZRange(ctx, key, start, stop)

	redisResults, err := command.Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		logger.Ctx(ctx).Warn("Error while ranging from sorted set",
			logger.String("key", key),
			logger.Error(err),
		)
	}

	results := make([][]byte, len(redisResults))

	for i, v := range redisResults {
		results[i] = []byte(v)
	}

	return results
}

// SAdd adds a member to a Redis set.
func (r *redisCache) SAdd(ctx context.Context, key string, val []byte, ttl time.Duration) bool {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	pipe := r.redisCacheCluster.Pipeline()

	pipe.SAdd(ctx, key, val)
	pipe.Expire(ctx, key, ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Ctx(ctx).Warn("Error while adding to set",
			logger.String("key", key),
			logger.Error(err),
		)

		return false
	}

	return true
}
