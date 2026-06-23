package invalidation

import "context"

type CacheGroupInvalidationConsumer struct {
	CacheType string   `json:"cache_type"`
	IDs       []string `json:"ids"`
}

func (c CacheGroupInvalidationConsumer) Execute(ctx context.Context) error {
	if c.CacheType != "" {
		deleteCacheGroup(ctx, c.CacheType, c.IDs, true)
	}

	return nil
}
