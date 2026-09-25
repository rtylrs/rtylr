package cacher

import "context"

type LocalCacheInvalidationConsumer struct {
	CacheKeys []string `json:"cache_model"`
}

func (c LocalCacheInvalidationConsumer) Execute(ctx context.Context) error {
	if len(c.CacheKeys) > 0 {
		GetInMemoryCache(ctx).Delete(ctx, c.CacheKeys)
	}

	return nil
}
