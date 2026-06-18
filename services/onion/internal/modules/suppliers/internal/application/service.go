package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/suppliers/internal/domain"
	"github.com/voxire/onion/internal/modules/suppliers/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const (
	supplierCacheTTL          = 5 * time.Minute
	suppliersListCacheVersion = "v2"
)

type repository interface {
	ListSuppliers(ctx context.Context, ownerID int64, limit, offset int) ([]domain.Supplier, bool, error)
	CreateSupplier(ctx context.Context, ownerID int64, input domain.SupplierInput) (int64, error)
	UpdateSupplier(ctx context.Context, ownerID, id int64, input domain.SupplierInput) error
	DeleteSupplier(ctx context.Context, ownerID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func paginateSuppliers(page, pageSize int64) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 500 {
		pageSize = 500
	}
	return int(pageSize), int((page - 1) * pageSize)
}

func (s *Service) ListSuppliers(ctx context.Context, ownerID, page, pageSize int64) ([]domain.Supplier, bool, error) {
	limit, offset := paginateSuppliers(page, pageSize)
	cacheKey := fmt.Sprintf("suppliers:%d|%s|p=%d|ps=%d", ownerID, suppliersListCacheVersion, page, pageSize)

	type pageResult struct {
		items   []domain.Supplier
		hasMore bool
	}

	res, err := cacher.Of[pageResult]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) (pageResult, error) {
		items, hasMore, err := s.repo.ListSuppliers(ctx, ownerID, limit, offset)
		if items == nil {
			items = []domain.Supplier{}
		}
		return pageResult{items: items, hasMore: hasMore}, err
	}, cacher.WithExpiry(ctx, supplierCacheTTL))
	if err != nil {
		return nil, false, err
	}
	if res.items == nil {
		res.items = []domain.Supplier{}
	}
	return res.items, res.hasMore, nil
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
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, fmt.Sprintf("*suppliers:%d*", ownerID))
	}
}
