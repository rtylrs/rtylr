package cacher

import (
	"context"
	"encoding/binary"
	"errors"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/allegro/bigcache/v3"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	entryNotFoundError = "Entry not found"
	expiryBytesLen     = 8
)

type LocalCache struct {
	cache *bigcache.BigCache
	mu    sync.Mutex
}

func (c *LocalCache) GetMulti(ctx context.Context, keys []string) (results map[string][]byte, notFound []string) {
	results = make(map[string][]byte, len(keys))

	for _, key := range keys {
		results[key] = c.Get(ctx, key)
		if results[key] == nil {
			notFound = append(notFound, key)
		}
	}

	return results, notFound
}

func (c *LocalCache) PutMulti(ctx context.Context, keyVals map[string][]byte, ttl time.Duration) {
	for key, val := range keyVals {
		if val != nil {
			GetInMemoryCache(ctx).Put(ctx, key, val, ttl)
		}
	}
}

func (c *LocalCache) Get(ctx context.Context, key string) []byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.get(ctx, key)
}

func (c *LocalCache) get(ctx context.Context, key string) []byte {
	if c.cache == nil {
		return nil // Gracefully handle uninitialized cache
	}

	result, err := c.cache.Get(key)
	if err != nil {
		if !errors.Is(err, bigcache.ErrEntryNotFound) && err.Error() != entryNotFoundError {
			logger.Ctx(ctx).Warn("Error while retrieving data from the memory cache",
				logger.String("key", key),
				logger.Error(err),
			)
		}

		return nil
	}

	if len(result) < expiryBytesLen {
		logger.Ctx(ctx).Warn("Invalid memory cache entry",
			logger.String("key", key),
		)

		return nil
	}

	expiresAt := int64(binary.BigEndian.Uint64(result[:expiryBytesLen])) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	if time.Now().UnixNano() > expiresAt {
		err = c.cache.Delete(key)

		if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) && err.Error() != entryNotFoundError {
			logger.Ctx(ctx).Warn("Error while deleting data from the memory cache",
				logger.String("key", key),
				logger.Error(err),
			)
		}

		return nil
	}

	return result[expiryBytesLen:]
}

// GetDel retrieves and removes a value from the local cache. BigCache performs
// the delete under its own lock, so one-time credentials cannot be replayed by
// concurrent requests within a process.
func (c *LocalCache) GetDel(ctx context.Context, key string) []byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	value := c.get(ctx, key)
	if value != nil {
		c.delete(ctx, []string{key})
	}
	return value
}

func (c *LocalCache) Put(ctx context.Context, key string, val []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.put(ctx, key, val, ttl)
}

func (c *LocalCache) put(ctx context.Context, key string, val []byte, ttl time.Duration) {
	if c.cache == nil {
		return // Gracefully handle uninitialized cache
	}

	entry := make([]byte, expiryBytesLen+len(val))
	binary.BigEndian.PutUint64(entry[:expiryBytesLen], uint64(time.Now().Add(ttl).UnixNano()))
	copy(entry[expiryBytesLen:], val)

	err := c.cache.Set(key, entry)
	if err != nil {
		logger.Ctx(ctx).Warn("Error while putting data in the memory cache",
			logger.String("key", key),
			logger.Error(err),
		)
	}
}

func (c *LocalCache) Delete(ctx context.Context, keys []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.delete(ctx, keys)
}

func (c *LocalCache) delete(ctx context.Context, keys []string) {
	if c.cache == nil {
		return // Gracefully handle uninitialized cache
	}

	for _, key := range keys {
		err := c.cache.Delete(key)

		if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) && err.Error() != entryNotFoundError {
			logger.Ctx(ctx).Warn("Error while deleting from memory cache",
				logger.Error(err),
				logger.String("keys", strings.Join(keys, ",")),
			)
		}
	}
}

func (c *LocalCache) DeleteByPattern(ctx context.Context, pattern string) int {
	if c.cache == nil || pattern == "" {
		return 0
	}

	patterns := cachePatternVariants(pattern)
	deleted := 0
	seen := make(map[string]struct{})
	iter := c.cache.Iterator()

	for iter.SetNext() {
		entry, err := iter.Value()
		if err != nil {
			logger.Ctx(ctx).Warn("Error while iterating memory cache",
				logger.Error(err),
				logger.String("pattern", pattern),
			)
			continue
		}

		key := entry.Key()
		if _, ok := seen[key]; ok {
			continue
		}

		for _, currentPattern := range patterns {
			matched, matchErr := path.Match(currentPattern, key)
			if matchErr != nil {
				logger.Ctx(ctx).Warn("Invalid memory cache delete pattern",
					logger.Error(matchErr),
					logger.String("pattern", currentPattern),
				)
				break
			}
			if !matched {
				continue
			}

			c.Delete(ctx, []string{key})
			seen[key] = struct{}{}
			deleted++
			break
		}
	}

	return deleted
}

var localCacheInstance *LocalCache

// InitializeLocalCache initializes the BigCache instance with proper error handling.
func InitializeLocalCache(ctx context.Context) error {
	bigCacheInstance, err := bigcache.New(ctx, getConfig())
	if err != nil {
		logger.Ctx(ctx).Error("Could not initialize bigcache", logger.Error(err))

		return err
	}

	localCacheInstance = &LocalCache{
		cache: bigCacheInstance,
	}

	logger.Ctx(ctx).Info("Local cache initialized successfully")

	return nil
}

func getConfig() bigcache.Config {
	const (
		defaultNumOfShards                = 128
		defaultLocalCacheEntryAge         = 45 * time.Minute
		defaultLocalCacheCleanUpPeriod    = 30 * time.Second
		defaultLocalCacheMaxEntries       = 1000 * 10 * 60
		defaultLocalCacheEntrySizeInBytes = 500
	)

	return bigcache.Config{
		// number of shards (must be a power of 2)
		Shards: defaultNumOfShards,

		// time after which entry can be evicted
		LifeWindow: defaultLocalCacheEntryAge,

		// Interval between removing expired entries (clean up).
		// If set to <= 0 then no action is performed.
		// Setting to < 1 second is counterproductive — bigcache has a one second resolution.
		CleanWindow: defaultLocalCacheCleanUpPeriod,

		// rps * lifeWindow, used only in initial memory allocation
		MaxEntriesInWindow: defaultLocalCacheMaxEntries,

		// max entry size in bytes, used only in initial memory allocation
		MaxEntrySize: defaultLocalCacheEntrySizeInBytes,

		// prints information about additional memory allocation
		Verbose: false,

		// cache will not allocate more memory than this limit, value in MB
		// if value is reached then the oldest entries can be overridden for the new ones
		// 0 value means no size limit
		HardMaxCacheSize: config.GetInMemoryCacheMaxSizeMb(),

		// callback fired when the oldest entry is removed because of its expiration time or no space left
		// for the new entry, or because delete was called. A bitmask representing the reason will be returned.
		// Default value is nil which means no callback and it prevents from unwrapping the oldest entry.
		OnRemove: nil,

		// OnRemoveWithReason is a callback fired when the oldest entry is removed because of its expiration time or no space left
		// for the new entry, or because delete was called. A constant representing the reason will be passed through.
		// Default value is nil which means no callback and it prevents from unwrapping the oldest entry.
		// Ignored if OnRemove is specified.
		OnRemoveWithReason: nil,
	}
}

func GetInMemoryCache(ctx context.Context) *LocalCache {
	if localCacheInstance == nil {
		logger.Ctx(ctx).Warn("Local cache not initialized, operations will be no-ops")

		return &LocalCache{} // Return empty instance that won't panic
	}

	return localCacheInstance
}

func GetLocalCache(ctx context.Context) *LocalCache {
	if localCacheInstance == nil {
		logger.Ctx(ctx).Warn("Local cache not initialized, operations will be no-ops")

		return &LocalCache{} // Return empty instance that won't panic
	}

	return localCacheInstance
}
