package cacher

import (
	"context"
	"errors"
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
	// PutConfirmed is the error-reporting form of Put for callers that must
	// not continue after a lost write (for example, one-time credentials).
	PutConfirmed(ctx context.Context, key string, val []byte, ttl time.Duration) bool
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
	// the earlier Redis setup.
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
