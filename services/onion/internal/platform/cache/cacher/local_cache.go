package cacher

import (
	"context"
	"strings"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	entryNotFoundError = "Entry not found"
)

type LocalCache struct {
	cache *bigcache.BigCache
}

func (c *LocalCache) GetMulti(ctx context.Context, keys []string) (results map[string][]byte, notFound []string) {
	results = make(map[string][]byte, len(keys))

	for _, key := range keys {
		results[key] = GetInMemoryCache(ctx).Get(ctx, key)
		if results[key] == nil {
			notFound = append(notFound, key)
		}
	}

	return
}

func (c *LocalCache) PutMulti(ctx context.Context, keyVals map[string][]byte, ttl time.Duration) {
	for key, val := range keyVals {
		if val != nil {
			GetInMemoryCache(ctx).Put(ctx, key, val, ttl)
		}
	}
}

func (c *LocalCache) Get(ctx context.Context, key string) []byte {
	if c.cache == nil {
		return nil // Gracefully handle uninitialized cache
	}

	int64BytesLen := 15

	result, err := c.cache.Get(key)
	if err != nil {
		if err.Error() != entryNotFoundError {
			logger.Ctx(ctx).Warn("Error while retrieving data from the memory cache",
				logger.String("key", key),
				logger.Error(err),
			)
		}

		return nil
	}

	var bestBefore time.Time

	err = bestBefore.UnmarshalBinary(result[:int64BytesLen])
	if err != nil {
		logger.Ctx(ctx).Warn("Error while unmarshaling data from the memory cache",
			logger.String("key", key),
			logger.Error(err),
		)

		return nil
	}

	if time.Now().After(bestBefore) {
		err = c.cache.Delete(key)

		if err != nil && err.Error() != entryNotFoundError {
			logger.Ctx(ctx).Warn("Error while deleting data from the memory cache",
				logger.String("key", key),
				logger.Error(err),
			)
		}

		return nil
	}

	return result[int64BytesLen:]
}

func (c *LocalCache) Put(ctx context.Context, key string, val []byte, ttl time.Duration) {
	if c.cache == nil {
		return // Gracefully handle uninitialized cache
	}

	expiryBytes := getExpiryBytes(ttl)

	err := c.cache.Set(key, append(expiryBytes, val...))
	if err != nil {
		logger.Ctx(ctx).Warn("Error while putting data in the memory cache",
			logger.String("key", key),
			logger.Error(err),
		)
	}
}

func (c *LocalCache) Delete(ctx context.Context, keys []string) {
	if c.cache == nil {
		return // Gracefully handle uninitialized cache
	}

	for _, key := range keys {
		err := c.cache.Delete(key)

		if err != nil && err.Error() != entryNotFoundError {
			logger.Ctx(ctx).Warn("Error while deleting from memory cache",
				logger.Error(err),
				logger.String("keys", strings.Join(keys, ",")),
			)
		}
	}
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

func getExpiryBytes(ttl time.Duration) []byte {
	timestamp := time.Now().Add(ttl)

	binaryTime, err := timestamp.MarshalBinary()
	if err != nil {
		// In case of errors default to an expiry on 01-01-2100 at midnight
		// Create a proper time.Time and marshal it to binary format
		fallbackTime := time.Unix(4102444800, 0)

		binaryTime, err = fallbackTime.MarshalBinary()
		if err != nil {
			// If even the fallback fails, use a far future time
			farFutureTime := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
			binaryTime, err = farFutureTime.MarshalBinary()
			// If this also fails, we'll use the zero value of []byte
			if err != nil {
				binaryTime = []byte{}
			}
		}
	}

	return binaryTime
}
