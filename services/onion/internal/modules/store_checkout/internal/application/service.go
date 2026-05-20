package application

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/store_checkout/internal/domain"
	"github.com/voxire/onion/internal/modules/store_checkout/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

type repository interface {
	Create(ctx context.Context, ownerID int64, input domain.CheckoutCreateInput) (int64, error)
	Complete(ctx context.Context, ownerID, id int64) error
}

type Service struct{ repo repository }

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

// store_checkout is write-only — no cached reads here. Mutations still wipe
// the order/customer scopes upstream so derived caches stay fresh.
func invalidateCheckoutScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.OrderPrefix(ownerID))
		cache.DeleteByPattern(ctx, cacheKeys.CustomerPrefix(ownerID))
	}
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.CheckoutCreateInput) (int64, error) {
	id, err := s.repo.Create(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}
	invalidateCheckoutScope(ctx, ownerID)
	return id, nil
}

func (s *Service) Complete(ctx context.Context, ownerID, id int64) error {
	if err := s.repo.Complete(ctx, ownerID, id); err != nil {
		return err
	}
	invalidateCheckoutScope(ctx, ownerID)
	return nil
}
