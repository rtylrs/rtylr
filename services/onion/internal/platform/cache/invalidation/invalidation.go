package invalidation

import (
	"context"
	"sort"
	"sync"

	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	UserCacheGroup = "user"
)

const (
	CacheGroupInvalidationChannel = "invalidate-cache-group"
)

type InvalidateFunc func(ctx context.Context, ids []string, options ...cacher.ClientOption)

var (
	mu              sync.RWMutex
	invalidationMap map[string]InvalidateFunc
)

func Initialize() {
	mu.Lock()
	defer mu.Unlock()
	invalidationMap = map[string]InvalidateFunc{}
}

func RegisterInvalidation(invMap map[string]InvalidateFunc) {
	RegisterGroups(invMap)
}

func RegisterGroups(invMap map[string]InvalidateFunc) {
	mu.Lock()
	defer mu.Unlock()

	if invalidationMap == nil {
		invalidationMap = map[string]InvalidateFunc{}
	}
	for route, handler := range invMap {
		if route == "" || handler == nil {
			continue
		}
		invalidationMap[route] = handler
	}
}

func AvailableGroups() []string {
	mu.RLock()
	defer mu.RUnlock()

	groups := make([]string, 0, len(invalidationMap))
	for group := range invalidationMap {
		groups = append(groups, group)
	}
	sort.Strings(groups)
	return groups
}

func InvalidateCacheGroup(ctx context.Context, cacheType string, ids []string) bool {
	if !deleteCacheGroup(ctx, cacheType, ids, false) {
		return false
	}

	cacher.NewPubSub[CacheGroupInvalidationConsumer]().Publish(CacheGroupInvalidationChannel,
		CacheGroupInvalidationConsumer{
			CacheType: cacheType,
			IDs:       ids,
		})
	return true
}

func deleteCacheGroup(ctx context.Context, cacheType string, ids []string, localOnly bool) bool {
	var clientOptions []cacher.ClientOption
	if localOnly {
		clientOptions = append(clientOptions, cacher.MemoryOnly())
	}

	mu.RLock()
	invalidateFunc, found := invalidationMap[cacheType]
	mu.RUnlock()
	if !found {
		logger.Ctx(ctx).Warn("unknown cache invalidation group",
			logger.String("cache_group", cacheType),
		)
		return false
	}

	invalidateFunc(ctx, ids, clientOptions...)
	return true
}

func SubscribeCacheGroupInvalidations(ctx context.Context) {
	cacher.NewPubSub[CacheGroupInvalidationConsumer]().SubscribeContext(ctx, CacheGroupInvalidationChannel)
}
