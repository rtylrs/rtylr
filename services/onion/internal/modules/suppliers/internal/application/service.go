package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/suppliers/internal/domain"
	"github.com/voxire/onion/internal/modules/suppliers/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const supplierCacheTTL = 5 * time.Minute

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) ListSuppliers(ctx context.Context, ownerID int64) ([]domain.Supplier, error) {
	cacheKey := fmt.Sprintf("suppliers:%d", ownerID)

	return cacher.Of[[]domain.Supplier]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.Supplier, error) {
		return s.repo.ListSuppliers(ctx, ownerID)
	}, cacher.WithExpiry(ctx, supplierCacheTTL))
}

func (s *Service) CreateSupplier(ctx context.Context, ownerID int64, input domain.SupplierInput) (int64, error) {
	id, err := s.repo.CreateSupplier(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}
	s.invalidateCache(ctx, ownerID)
	return id, nil
}

func (s *Service) UpdateSupplier(ctx context.Context, ownerID, id int64, input domain.SupplierInput) error {
	err := s.repo.UpdateSupplier(ctx, ownerID, id, input)
	if err != nil {
		return err
	}
	s.invalidateCache(ctx, ownerID)
	return nil
}

func (s *Service) DeleteSupplier(ctx context.Context, ownerID, id int64) error {
	if err := s.repo.DeleteSupplier(ctx, ownerID, id); err != nil {
		return err
	}
	s.invalidateCache(ctx, ownerID)
	return nil
}

func (s *Service) invalidateCache(ctx context.Context, ownerID int64) {
	cacheKey := fmt.Sprintf("suppliers:%d", ownerID)
	cacher.Of[[]domain.Supplier]().Invalidate(ctx, []string{cacheKey})
}
