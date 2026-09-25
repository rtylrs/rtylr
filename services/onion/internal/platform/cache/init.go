package cache

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/cache/invalidation"
	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

var (
	ErrLocalCacheTestFailed = errors.New("local cache test failed: value mismatch")
	ErrRedisCacheTestFailed = errors.New("redis cache test failed: value mismatch")
)

// Initialize sets up the cache system with health checks and proper initialization.
func Initialize(ctx context.Context) error {
	logger.Ctx(ctx).Info("Initializing cache system...")

	// Initialize local cache (BigCache) first
	if err := cacher.InitializeLocalCache(ctx); err != nil {
		logger.Ctx(ctx).Error("Failed to initialize local cache", logger.Error(err))

		return err
	}

	// Initialize cache invalidation system
	invalidation.Initialize()

	// Test local cache (BigCache)
	if err := testLocalCache(ctx); err != nil {
		logger.Ctx(ctx).Error("Local cache initialization failed", logger.Error(err))

		return err
	}

	logger.Ctx(ctx).Info("Local cache initialized successfully")

	// Test Redis cache if enabled and configured
	if config.IsRedisConfigured() {
		if err := testRedisCache(ctx); err != nil {
			logger.Ctx(ctx).Warn("Redis cache initialization failed, using local cache only", logger.Error(err))
			cacher.DisableRedisForProcess(ctx)
		} else {
			logger.Ctx(ctx).Info("Redis cache initialized successfully")
		}
	} else if config.IsRedisEnabled() {
		logger.Ctx(ctx).Warn("Redis is enabled but not properly configured, using local cache only")
	} else {
		logger.Ctx(ctx).Info("Redis disabled, using local cache only")
	}

	logger.Ctx(ctx).Info("Cache system initialization completed successfully")

	return nil
}

func RedisAvailable() bool {
	return cacher.RedisAvailable()
}

// testLocalCache performs a basic health check on the local cache.
func testLocalCache(ctx context.Context) error {
	testKey := "cache-init-test"
	testValue := []byte("test-value")
	ttl := 1 * time.Second

	// Get local cache instance
	localCache := cacher.GetLocalCache(ctx)

	// Test set operation
	localCache.Put(ctx, testKey, testValue, ttl)

	// Test get operation
	retrieved := localCache.Get(ctx, testKey)

	// Verify the value
	if !bytes.Equal(retrieved, testValue) {
		logger.Ctx(ctx).Error("Local cache test failed: value mismatch",
			logger.String("expected", string(testValue)),
			logger.String("actual", string(retrieved)),
		)

		return ErrLocalCacheTestFailed
	}

	// Clean up test key
	localCache.Delete(ctx, []string{testKey})

	return nil
}

// testRedisCache performs a basic health check on the Redis cache.
func testRedisCache(ctx context.Context) error {
	// Multiple API/worker/upload processes can initialize concurrently. A
	// shared probe key lets one process delete another process's value and
	// falsely report a mismatch, which used to produce noisy ERROR events and
	// unnecessarily disable Redis for a healthy process.
	testKey := fmt.Sprintf("cache-init-test-redis:%d", time.Now().UnixNano())
	testValue := []byte("test-value-redis")
	ttl := 1 * time.Second

	// Get Redis cache instance
	redisCache := cacher.GetRedisCache()

	// Test set operation
	redisCache.Put(ctx, testKey, testValue, ttl)

	// Test get operation
	retrieved := redisCache.Get(ctx, testKey)

	// Verify the value
	if !bytes.Equal(retrieved, testValue) {
		logger.Ctx(ctx).Error("Redis cache test failed: value mismatch",
			logger.String("expected", string(testValue)),
			logger.String("actual", string(retrieved)),
		)

		return ErrRedisCacheTestFailed
	}

	// Clean up test key
	redisCache.Delete(ctx, []string{testKey})

	return nil
}

// GetCacheStatus returns the current status of both cache systems.
func GetCacheStatus(ctx context.Context) Status {
	status := Status{
		LocalCache: SystemStatus{
			Name:    "Local Cache (BigCache)",
			Enabled: true,
		},
		RedisCache: SystemStatus{
			Name:    "Redis Cache",
			Enabled: config.IsRedisConfigured(),
		},
	}

	// Test local cache
	if err := testLocalCache(ctx); err != nil {
		status.LocalCache.Status = "unhealthy"
		status.LocalCache.Error = err.Error()
	} else {
		status.LocalCache.Status = "healthy"
	}

	// Test Redis cache if enabled and configured
	if config.IsRedisConfigured() {
		if err := testRedisCache(ctx); err != nil {
			status.RedisCache.Status = "unhealthy"
			status.RedisCache.Error = err.Error()
		} else {
			status.RedisCache.Status = "healthy"
		}
	} else if config.IsRedisEnabled() {
		status.RedisCache.Status = "misconfigured"
		status.RedisCache.Error = "Redis enabled but not properly configured"
	} else {
		status.RedisCache.Status = "disabled"
	}

	return status
}

// Status represents the overall status of the cache system.
type Status struct {
	LocalCache SystemStatus `json:"local_cache"`
	RedisCache SystemStatus `json:"redis_cache"`
}

// SystemStatus represents the status of an individual cache system.
type SystemStatus struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Status  string `json:"status"` // "healthy", "unhealthy", "disabled", "misconfigured"
	Error   string `json:"error,omitempty"`
}
