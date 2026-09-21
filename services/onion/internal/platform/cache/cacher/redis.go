package cacher

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/maputils"
)

// RedisCache defines the interface for Redis cache operations.
type RedisCache interface {
	Put(ctx context.Context, key string, val []byte, ttl time.Duration)
	Get(ctx context.Context, key string) []byte
	// GetDel atomically gets and deletes a key. It is used for one-time
	// credentials where a separate Get followed by Delete would allow replay.
	GetDel(ctx context.Context, key string) []byte
	PutMulti(ctx context.Context, keyVals map[string][]byte, ttl time.Duration)
	GetMulti(ctx context.Context, keys []string) (results map[string][]byte, notFound []string)
	Delete(ctx context.Context, keys []string)
	DeleteByPattern(ctx context.Context, pattern string) int
	HSet(ctx context.Context, key, field string, val []byte, ttl time.Duration)
	HFieldExpire(ctx context.Context, key, field string, ttl time.Duration)
	HSetWithFieldExpiry(ctx context.Context, key, field string, val []byte, fieldExpiry, ttl time.Duration)
	HDel(ctx context.Context, key, field string) bool
	HGet(ctx context.Context, key, field string) []byte
	HGetAll(ctx context.Context, key string) map[string][]byte
	ZAdd(ctx context.Context, key string, val []byte, score float64, ttl time.Duration) bool
	ZRem(ctx context.Context, key string, val []byte) bool
	ZRemRangeUpToScore(ctx context.Context, key string, upToScore float64) bool
	ZRemRangeByRank(ctx context.Context, key string, start, end int64) bool
	ZRange(ctx context.Context, key string, start, end int64) [][]byte
	ZCard(ctx context.Context, key string) (int64, bool)
	SAdd(ctx context.Context, key string, val []byte, ttl time.Duration) bool
	SRem(ctx context.Context, key string, val []byte) bool
	SMembers(ctx context.Context, key string) [][]byte
	SCard(ctx context.Context, key string) (int64, bool)
	HSetNx(ctx context.Context, key, field string, val []byte, ttl time.Duration)
	Increment(ctx context.Context, key string, ttl time.Duration) int64
	GetInt(ctx context.Context, key string) []byte
	SetNx(ctx context.Context, key string, val []byte, ttl time.Duration) bool
	Expire(ctx context.Context, key string, ttl time.Duration)
	ExpireAt(ctx context.Context, key string, ttl time.Duration)
}

type redisCache struct {
	redisCacheCluster redis.UniversalClient
}

var (
	client             redis.UniversalClient
	redisCacheInstance RedisCache
)

func getTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	const defaultRedisTimeout = 3 * time.Second

	return context.WithTimeout(ctx, defaultRedisTimeout)
}

func init() {
	if !config.IsRedisEnabled() {
		redisCacheInstance = &noopRedisCache{}
		return
	}

	client = getNewRedisClient()
	redisCacheInstance = &redisCache{
		redisCacheCluster: client,
	}
}

func getNewRedisClient() redis.UniversalClient {
	const (
		defaultMinClusterClientIdleConns = 2
		defaultPoolSize                  = 25
	)

	// Prefer REDIS_URL (Upstash / Heroku style). Falls back to the
	// REDIS_HOSTS / REDIS_USERNAME / REDIS_PASSWORD env trio used by
	// our older self-hosted setup.
	if url := config.GetRedisURL(); url != "" {
		opt, err := redis.ParseURL(url)
		if err == nil {
			opt.ClientName = "onion-api"
			opt.DisableIdentity = true
			opt.PoolSize = defaultPoolSize
			opt.MinIdleConns = defaultMinClusterClientIdleConns
			return redis.NewClient(opt)
		}
		logger.Global().Warn("REDIS_URL failed to parse; falling back to REDIS_HOSTS", logger.Error(err))
	}

	redisClient := redis.NewUniversalClient(&redis.UniversalOptions{
		ClientName:      "onion-api",
		Addrs:           config.GetRedisHosts(),
		Username:        config.GetRedisUsername(),
		Password:        config.GetRedisPassword(),
		DisableIdentity: true,
		PoolSize:        defaultPoolSize,
		MinIdleConns:    defaultMinClusterClientIdleConns,
	})

	return redisClient
}

func getRedisCache() RedisCache {
	return redisCacheInstance
}

func GetRedisCache() RedisCache {
	return redisCacheInstance
}

func DisableRedisForProcess(ctx context.Context) {
	if client != nil {
		if err := client.Close(); err != nil {
			logger.Ctx(ctx).Warn("failed to close redis client", logger.Error(err))
		}
	}
	client = nil
	redisCacheInstance = &noopRedisCache{}
}

func RedisAvailable() bool {
	return client != nil
}

// PutMulti stores multiple key-value pairs in Redis.
func (r *redisCache) PutMulti(ctx context.Context, keyVals map[string][]byte, ttl time.Duration) {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	pipe := r.redisCacheCluster.Pipeline()
	for key, value := range keyVals {
		pipe.Set(ctx, key, string(value), ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Ctx(ctx).Warn("Error while putting data in redis",
			logger.String("keys", strings.Join(maputils.Keys(keyVals), ",")),
			logger.Error(err),
		)
	}
}

// GetMulti retrieves the values associated with multiple keys from Redis.
// It takes a slice of keys as input and returns a map of key-value pairs
// representing the values found in Redis for the provided keys.
func (r *redisCache) GetMulti(ctx context.Context, keys []string) (results map[string][]byte, notFound []string) {
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	if len(keys) == 0 {
		results = make(map[string][]byte)

		return results, notFound
	}

	results = make(map[string][]byte, len(keys))

	pipe := r.redisCacheCluster.Pipeline()

	for _, key := range keys {
		pipe.Get(ctx, key)
	}

	redisResults, err := pipe.Exec(ctx)

	if err != nil && !errors.Is(err, redis.Nil) {
		key := strings.Join(keys, ",")
		logRedisError(ctx, "redis error in GetMulti retrieving data", key, "", err)

		notFound = keys

		return results, notFound
	}

	for k, v := range redisResults {
		cmd, ok := v.(*redis.StringCmd)
		if !ok {
			continue
		}

		vVal := cmd.Val()

		if v == nil || (v.Err() != nil && errors.Is(v.Err(), redis.Nil)) {
			results[keys[k]] = nil
			notFound = append(notFound, keys[k])

			continue
		}

		results[keys[k]] = []byte(vVal)
	}

	return results, notFound
}

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
