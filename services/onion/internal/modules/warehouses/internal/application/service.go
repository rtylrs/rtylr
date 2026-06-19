package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/warehouses/internal/domain"
	"github.com/voxire/onion/internal/modules/warehouses/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const warehousesCacheTTL = 5 * time.Minute

type repository interface {
	ListWarehouses(ctx context.Context, businessID int64) ([]domain.Warehouse, error)
	CreateWarehouse(ctx context.Context, businessID int64, input domain.WarehouseInput) (int64, error)
	UpdateWarehouse(ctx context.Context, businessID, id int64, input domain.WarehouseInput) error
	DeleteWarehouse(ctx context.Context, businessID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	repo := infrastructure.NewRepository(db)

	return &Service{
		repo: repo,
	}
}

func (s *Service) ListWarehouses(ctx context.Context, businessID int64) ([]domain.Warehouse, error) {
	cacheKey := fmt.Sprintf("warehouses:%d", businessID)

	return cacher.Of[[]domain.Warehouse]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.Warehouse, error) {
		return s.repo.ListWarehouses(ctx, businessID)
	}, cacher.WithExpiry(ctx, warehousesCacheTTL))
}

func (s *Service) CreateWarehouse(ctx context.Context, businessID int64, input domain.WarehouseInput) (int64, error) {
	id, err := s.repo.CreateWarehouse(ctx, businessID, input)
	if err != nil {
		return 0, err
	}

	s.invalidateCache(ctx, businessID)

	return id, nil
}

func (s *Service) UpdateWarehouse(ctx context.Context, businessID, id int64, input domain.WarehouseInput) error {
	err := s.repo.UpdateWarehouse(ctx, businessID, id, input)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, businessID)

	return nil
}

func (s *Service) DeleteWarehouse(ctx context.Context, businessID, id int64) error {
	err := s.repo.DeleteWarehouse(ctx, businessID, id)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, businessID)

	return nil
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	prefix := fmt.Sprintf("warehouses:%d", businessID)
	cacher.Of[[]domain.Warehouse]().Invalidate(ctx, []string{prefix})
}
