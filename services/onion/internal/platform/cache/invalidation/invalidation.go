package invalidation

import (
	"context"
	"errors"
	"fmt"
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

// ErrUnknownCacheGroup is returned when a caller asks for a group that has not
// been registered. Keeping this distinct lets the HTTP layer preserve its
// existing 400 response while still surfacing validation errors from a
// registered group.
var ErrUnknownCacheGroup = errors.New("unknown cache invalidation group")

// InvalidateFunc invalidates one cache group. Implementations must return an
// error when the request is malformed instead of silently treating it as a
// successful no-op; the public endpoint uses that error to avoid reporting an
// invalidation that never happened as OK.
type InvalidateFunc func(ctx context.Context, ids []string, options ...cacher.ClientOption) error

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

// InvalidateCacheGroup preserves the original boolean API for callers that
// only need success/failure. New request handlers should use
// InvalidateCacheGroupWithError so malformed registered groups are not hidden.
func InvalidateCacheGroup(ctx context.Context, cacheType string, ids []string) bool {
	return InvalidateCacheGroupWithError(ctx, cacheType, ids) == nil
}

// InvalidateCacheGroupWithError performs the local invalidation and publishes
// the cross-process invalidation only after the registered group accepts the
// request. Publishing an invalid request would make every subscriber repeat a
// silent no-op.
func InvalidateCacheGroupWithError(ctx context.Context, cacheType string, ids []string) error {
	if err := deleteCacheGroup(ctx, cacheType, ids, false); err != nil {
		return err
	}

	cacher.NewPubSub[CacheGroupInvalidationConsumer]().Publish(CacheGroupInvalidationChannel,
		CacheGroupInvalidationConsumer{
			CacheType: cacheType,
			IDs:       ids,
		})
	return nil
}

func deleteCacheGroup(ctx context.Context, cacheType string, ids []string, localOnly bool) error {
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
		return fmt.Errorf("%w: %s", ErrUnknownCacheGroup, cacheType)
	}

	return invalidateFunc(ctx, ids, clientOptions...)
}

func SubscribeCacheGroupInvalidations(ctx context.Context) {
	cacher.NewPubSub[CacheGroupInvalidationConsumer]().SubscribeContext(ctx, CacheGroupInvalidationChannel)
}
