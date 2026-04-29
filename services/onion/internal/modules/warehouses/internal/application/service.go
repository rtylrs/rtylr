package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/warehouses/internal/domain"
	"github.com/voxire/onion/internal/modules/warehouses/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const warehousesCacheTTL = 5 * time.Minute

type repository interface {
	ListWarehouses(ctx context.Context, ownerID int64) ([]domain.Warehouse, error)
	CreateWarehouse(ctx context.Context, ownerID int64, input domain.WarehouseInput) (int64, error)
	UpdateWarehouse(ctx context.Context, ownerID, id int64, input domain.WarehouseInput) error
	DeleteWarehouse(ctx context.Context, ownerID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	repo := infrastructure.NewRepository(ctx)

	return &Service{
		repo: repo,
	}
}

func (s *Service) ListWarehouses(ctx context.Context, ownerID int64) ([]domain.Warehouse, error) {
	cacheKey := fmt.Sprintf("warehouses:%d", ownerID)

	return cacher.Of[[]domain.Warehouse]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.Warehouse, error) {
		return s.repo.ListWarehouses(ctx, ownerID)
	}, cacher.WithExpiry(ctx, warehousesCacheTTL))
}

func (s *Service) CreateWarehouse(ctx context.Context, ownerID int64, input domain.WarehouseInput) (int64, error) {
	id, err := s.repo.CreateWarehouse(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}

	s.invalidateCache(ctx, ownerID)

	return id, nil
}

func (s *Service) UpdateWarehouse(ctx context.Context, ownerID, id int64, input domain.WarehouseInput) error {
	err := s.repo.UpdateWarehouse(ctx, ownerID, id, input)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, ownerID)

	return nil
}

func (s *Service) DeleteWarehouse(ctx context.Context, ownerID, id int64) error {
	err := s.repo.DeleteWarehouse(ctx, ownerID, id)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, ownerID)

	return nil
}

func (s *Service) invalidateCache(ctx context.Context, ownerID int64) {
	prefix := fmt.Sprintf("warehouses:%d", ownerID)
	cacher.Of[[]domain.Warehouse]().Invalidate(ctx, []string{prefix})
}
