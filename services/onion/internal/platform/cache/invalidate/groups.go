package invalidate

import (
	"context"
	"fmt"

	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/cache/invalidation"
	"github.com/voxire/onion/internal/platform/cache/keys"
)

// RegisterDefaultGroups exposes the canonical pattern invalidators as named
// cache groups. Modules can still register narrower groups, but this gives
// /invalidate-cache a useful baseline: user, business, product, config, etc.
func RegisterDefaultGroups() {
	invalidation.RegisterGroups(map[string]invalidation.InvalidateFunc{
		string(keys.InvalidateUser):         patternGroup(keys.InvalidateUser),
		string(keys.InvalidateBusiness):     patternGroup(keys.InvalidateBusiness),
		string(keys.InvalidateOrder):        patternGroup(keys.InvalidateOrder),
		string(keys.InvalidateProduct):      patternGroup(keys.InvalidateProduct),
		string(keys.InvalidateCustomer):     patternGroup(keys.InvalidateCustomer),
		string(keys.InvalidateMenu):         patternGroup(keys.InvalidateMenu),
		string(keys.InvalidateTeam):         patternGroup(keys.InvalidateTeam),
		string(keys.InvalidateConfig):       patternGroup(keys.InvalidateConfig),
		string(keys.InvalidateSubscription): patternGroup(keys.InvalidateSubscription),
		string(keys.InvalidateAll):          patternGroup(keys.InvalidateAll),
	})
}

func patternGroup(t keys.InvalidationType) invalidation.InvalidateFunc {
	return func(ctx context.Context, ids []string, _ ...cacher.ClientOption) error {
		if len(ids) == 0 {
			if t == keys.InvalidateAll {
				Invalidate(ctx, t, "")
				return nil
			}
			return fmt.Errorf("ids are required for %s cache invalidation", t)
		}

		for _, id := range ids {
			result := Invalidate(ctx, t, id)
			if len(result.Patterns) == 0 {
				return fmt.Errorf("id %q is invalid for %s cache invalidation", id, t)
			}
		}
		return nil
	}
}
