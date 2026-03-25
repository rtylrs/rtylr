package invalidation

import (
	"context"
	"slices"

	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/sliceutils"
)

const (
	UserCacheGroup = "user"
)

const (
	CacheGroupInvalidationChannel = "invalidate-cache-group"
)

var availableGroups = []string{
	UserCacheGroup,
}

type InvalidateFunc func(ctx context.Context, ids []string, options ...cacher.ClientOption)

var invalidationMap map[string]InvalidateFunc

func Initialize() {
	invalidationMap = map[string]InvalidateFunc{}
}

func RegisterInvalidation(invMap map[string]InvalidateFunc) {
	for route, handler := range invMap {
		invalidationMap[route] = handler
	}
}

func InvalidateCacheGroup(ctx context.Context, cacheType string, ids []string) {
	if !sliceutils.Contains(availableGroups, cacheType) {
		return
	}

	deleteCacheGroup(ctx, cacheType, ids, false)

	cacher.NewPubSub[CacheGroupInvalidationConsumer]().Publish(CacheGroupInvalidationChannel,
		CacheGroupInvalidationConsumer{
			CacheType: cacheType,
			IDs:       ids,
		})
}

func deleteCacheGroup(ctx context.Context, cacheType string, ids []string, localOnly bool) {
	var clientOptions []cacher.ClientOption
	if localOnly {
		clientOptions = append(clientOptions, cacher.MemoryOnly())
	}

	// Look up the invalidation function based on the cache type
	if invalidateFunc, found := invalidationMap[cacheType]; found {
		invalidateFunc(ctx, ids, clientOptions...)
	} else {
		logger.Ctx(ctx).Error("Failed to get invalidation method for cache group",
			logger.String("cache_group", cacheType),
			logger.Bool("is_available_group", slices.Contains(availableGroups, cacheType)),
		)
	}
}
