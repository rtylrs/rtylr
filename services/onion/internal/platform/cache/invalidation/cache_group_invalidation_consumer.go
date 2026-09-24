package invalidation

import (
	"context"

	"github.com/voxire/onion/internal/platform/logger"
)

type CacheGroupInvalidationConsumer struct {
	CacheType string   `json:"cache_type"`
	IDs       []string `json:"ids"`
}

func (c CacheGroupInvalidationConsumer) Execute(ctx context.Context) error {
	if c.CacheType != "" {
		if err := deleteCacheGroup(ctx, c.CacheType, c.IDs, true); err != nil {
			logger.Ctx(ctx).Warn("cache group invalidation skipped",
				logger.String("cache_group", c.CacheType),
				logger.Error(err),
			)
		}
	}

	return nil
}
