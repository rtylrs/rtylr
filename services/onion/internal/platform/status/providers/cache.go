package providers

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/cache"
	"github.com/voxire/onion/internal/platform/status"
)

const healthyStatus = "healthy"

type CacheStatusProvider struct{}

func NewCacheStatusProvider() *CacheStatusProvider {
	return &CacheStatusProvider{}
}

func (c *CacheStatusProvider) GetName() string {
	return "cache"
}

func (c *CacheStatusProvider) GetStatus(ctx context.Context) status.ComponentStatus {
	cacheStatus := cache.GetCacheStatus(ctx)

	overallHealthy := cacheStatus.LocalCache.Status == healthyStatus &&
		(cacheStatus.RedisCache.Status == healthyStatus || cacheStatus.RedisCache.Status == "disabled")

	statusStr := healthyStatus
	message := "Cache system is operating normally"

	if !overallHealthy {
		statusStr = "unhealthy"
		message = "Cache system has issues"
	}

	details := map[string]interface{}{
		"local_cache": cacheStatus.LocalCache,
		"redis_cache": cacheStatus.RedisCache,
	}

	return status.ComponentStatus{
		Name:        "Cache",
		Status:      statusStr,
		Message:     message,
		Details:     details,
		LastChecked: time.Now().UTC().Format(time.RFC3339),
	}
}
