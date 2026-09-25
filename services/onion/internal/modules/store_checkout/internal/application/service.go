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
	Create(ctx context.Context, businessID, ownerID int64, input domain.CheckoutCreateInput) (int64, error)
	Complete(ctx context.Context, businessID, id, ownerID int64) error
}

type Service struct{ repo repository }

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

// store_checkout is write-only — no cached reads here. Mutations still wipe
// the order/customer scopes upstream so derived caches stay fresh.
func invalidateCheckoutScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx,
		cacheKeys.OrderPrefix(businessID),
		cacheKeys.CustomerPrefix(businessID),
	)
}

func (s *Service) Create(ctx context.Context, businessID, ownerID int64, input domain.CheckoutCreateInput) (int64, error) {
	id, err := s.repo.Create(ctx, businessID, ownerID, input)
	if err != nil {
		return 0, err
	}
	invalidateCheckoutScope(ctx, businessID)
	return id, nil
}

func (s *Service) Complete(ctx context.Context, businessID, id, ownerID int64) error {
	if err := s.repo.Complete(ctx, businessID, id, ownerID); err != nil {
		return err
	}
	invalidateCheckoutScope(ctx, businessID)
	return nil
}
