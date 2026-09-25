package cacher

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/voxire/onion/internal/platform/logger"
)

// Get retrieves the value associated with a single key from Redis.
func (r *redisCache) Get(ctx context.Context, key string) []byte {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	command := r.redisCacheCluster.Get(ctx, key)
	result, err := command.Bytes()

	if err != nil && !errors.Is(err, redis.Nil) {
		logRedisError(ctx, "redis error in Get retrieving data", key, "", err)
	}

	return result
}

// GetDel atomically retrieves and removes a key from Redis.
func (r *redisCache) GetDel(ctx context.Context, key string) []byte {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	command := r.redisCacheCluster.GetDel(ctx, key)
	result, err := command.Bytes()
	if err != nil && !errors.Is(err, redis.Nil) {
		logRedisError(ctx, "redis error in GetDel retrieving data", key, "", err)
	}

	return result
}

func (r *redisCache) GetInt(ctx context.Context, key string) []byte {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	command := r.redisCacheCluster.Get(ctx, key)
	result, err := command.Bytes()

	if err != nil && !errors.Is(err, redis.Nil) {
		logRedisError(ctx, "redis error in GetInt retrieving data", key, "", err)
	}

	return result
}

func logRedisError(ctx context.Context, msg, key, hashField string, err error) {
	if !errors.Is(err, context.Canceled) {
		if hashField != "" {
			logger.Ctx(ctx).Warn(msg,
				logger.String("key", key),
				logger.String("field", hashField),
				logger.Error(err),
			)
		} else {
			logger.Ctx(ctx).Warn(msg,
				logger.String("key", key),
				logger.Error(err),
			)
		}
	}
}

// Put stores a key-value pair in Redis.
func (r *redisCache) Put(ctx context.Context, key string, val []byte, ttl time.Duration) {
	r.PutConfirmed(ctx, key, val, ttl)
}

// PutConfirmed stores a key-value pair in Redis and reports whether the write
// succeeded. Callers that must not mistake a lost write for success (the
// idempotency reservation) use this instead of Put, which cannot report
// failure.
func (r *redisCache) PutConfirmed(ctx context.Context, key string, val []byte, ttl time.Duration) bool {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	status := r.redisCacheCluster.Set(ctx, key, val, ttl)

	redisError := status.Err()
	if redisError != nil {
		logger.Ctx(ctx).Warn("Error while putting data in redis",
			logger.String("key", key),
			logger.Error(redisError),
		)

		return false
	}

	return true
}

// HSet sets the value of a field in a Redis hash set.
func (r *redisCache) HSet(ctx context.Context, key, field string, val []byte, ttl time.Duration) {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	pipe := r.redisCacheCluster.Pipeline()
	pipe.HSet(ctx, key, field, val)
	pipe.Expire(ctx, key, ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Ctx(ctx).Warn("Error while putting data in redis",
			logger.String("key", key),
			logger.String("field", field),
			logger.Error(err),
		)
	}
}

// HFieldExpire sets expiry for a field in a hash set.
func (r *redisCache) HFieldExpire(ctx context.Context, key, field string, ttl time.Duration) {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	command := r.redisCacheCluster.HExpire(ctx, key, ttl, field)

	if command.Err() != nil {
		logger.Ctx(ctx).Warn("Error while HFieldExpire in redis",
			logger.String("key", key),
			logger.String("field", field),
			logger.Any("ttl", ttl),
			logger.Error(command.Err()),
		)
	}
}

// HSetWithFieldExpiry sets the value of a field with an expiry to the field in a Redis hash set.
func (r *redisCache) HSetWithFieldExpiry(ctx context.Context, key, field string, val []byte, fieldTTL, ttl time.Duration) {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	pipe := r.redisCacheCluster.Pipeline()
	pipe.HSet(ctx, key, field, val)
	pipe.HExpire(ctx, key, fieldTTL, field)
	pipe.Expire(ctx, key, ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Ctx(ctx).Warn("Error while putting data in redis hash set with expiry",
			logger.String("key", key),
			logger.String("field", field),
			logger.Error(err),
		)
	}
}

// HDel deletes a field from a Redis hash set.
func (r *redisCache) HDel(ctx context.Context, key, field string) bool {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	command := r.redisCacheCluster.HDel(ctx, key, field)

	if command.Err() != nil {
		logger.Ctx(ctx).Warn("Error while deleting from hash set",
			logger.String("key", key),
			logger.String("field", field),
			logger.Error(command.Err()),
		)

		return false
	}

	return true
}

// HGet retrieves the value of a field from a Redis hash set.
func (r *redisCache) HGet(ctx context.Context, key, field string) []byte {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	command := r.redisCacheCluster.HGet(ctx, key, field)
	result, err := command.Bytes()

	if err != nil && !errors.Is(err, redis.Nil) {
		logRedisError(ctx, "redis error in HGet retrieving data", key, field, err)
	}

	return result
}

// HGetAll retrieves all fields and values from a Redis hash set.
func (r *redisCache) HGetAll(ctx context.Context, key string) map[string][]byte {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	if key == "" {
		return make(map[string][]byte)
	}

	command := r.redisCacheCluster.HGetAll(ctx, key)
	redisResults, err := command.Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		logRedisError(ctx, "redis error in HGetAll retrieving data", key, "", err)
	}

	results := make(map[string][]byte, len(redisResults))

	for k, v := range redisResults {
		if v == "" {
			results[k] = nil

			continue
		}

		results[k] = []byte(v)
	}

	return results
}
