package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/store/payments/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	cacheTTL           = 5 * time.Minute
	storePaymentsScope = "store_payments"
)

type Service struct{ repo repository }

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.PaymentMethod, error) {
	cacheKey := cacheKeys.Scoped(storePaymentsScope, businessID, "methods")
	return cacher.Of[[]domain.PaymentMethod]().GetOrComputeSingle(ctx, cacheKey+"|domain=v2", func(ctx context.Context) ([]domain.PaymentMethod, error) {
		return s.repo.List(ctx, businessID)
	}, cacher.WithExpiry(ctx, cacheTTL))
}

func (s *Service) Create(ctx context.Context, ownerID, businessID int64, input domain.PaymentMethodInput) (int64, error) {
	id, err := s.repo.Create(ctx, ownerID, businessID, input)
	if err != nil {
		return 0, err
	}
	s.invalidateCache(ctx, businessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.PaymentMethodInput) error {
	if err := s.repo.Update(ctx, businessID, id, input); err != nil {
		return err
	}
	s.invalidateCache(ctx, businessID)
	return nil
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	cacher.Of[[]domain.PaymentMethod]().Invalidate(ctx, []string{cacheKeys.Scoped(storePaymentsScope, businessID, "methods") + "|domain=v2"})
}

type repository interface {
	Create(ctx context.Context, ownerID, businessID int64, input domain.PaymentMethodInput) (int64, error)
	List(ctx context.Context, businessID int64) ([]domain.PaymentMethod, error)
	Update(ctx context.Context, businessID, id int64, input domain.PaymentMethodInput) error
}
