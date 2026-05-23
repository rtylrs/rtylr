package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/inventory/internal/domain"
	"github.com/voxire/onion/internal/modules/inventory/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
)

const inventoryCacheTTL = 5 * time.Minute

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	repo := infrastructure.NewRepository(db)

	return &Service{
		repo: repo,
	}
}

// Inventory views live under the product namespace because they are derived
// from product stock — wiping ProductPrefix covers every inventory view too.
func invalidateInventoryScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ProductPrefix(ownerID))
	}
}

func (s *Service) AdjustInventory(ctx context.Context, ownerID int64, input domain.InventoryAdjustInput) error {
	if err := s.repo.AdjustInventory(ctx, ownerID, input.ProductID, input.Change, input.Type, input.Notes); err != nil {
		return err
	}

	invalidateInventoryScope(ctx, ownerID)

	lowStock, lowErr := s.repo.GetLowStockProducts(ctx, ownerID)
	if lowErr != nil {
		logger.Ctx(ctx).Error("Failed to check low stock after adjustment", logger.Error(lowErr))
	} else if len(lowStock) > 0 {
		logger.Ctx(ctx).Info("Low stock detected after inventory adjustment",
			logger.Any("owner_id", ownerID),
			logger.Any("low_stock_count", len(lowStock)),
		)
	}

	return nil
}

func (s *Service) GetInventoryReport(ctx context.Context, ownerID int64) ([]domain.InventoryReportItem, error) {
	return cacher.Of[[]domain.InventoryReportItem]().GetOrComputeSingle(ctx,
		cacheKeys.InventoryReport(ownerID),
		func(ctx context.Context) ([]domain.InventoryReportItem, error) {
			return s.repo.GetInventoryReport(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, inventoryCacheTTL),
	)
}

func (s *Service) GetLowStockProducts(ctx context.Context, ownerID int64) ([]domain.InventoryReportItem, error) {
	return cacher.Of[[]domain.InventoryReportItem]().GetOrComputeSingle(ctx,
		cacheKeys.InventoryLowStock(ownerID),
		func(ctx context.Context) ([]domain.InventoryReportItem, error) {
			return s.repo.GetLowStockProducts(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, inventoryCacheTTL),
	)
}

func (s *Service) GetValuation(ctx context.Context, ownerID int64) ([]domain.ValuationItem, error) {
	return cacher.Of[[]domain.ValuationItem]().GetOrComputeSingle(ctx,
		cacheKeys.InventoryValuation(ownerID),
		func(ctx context.Context) ([]domain.ValuationItem, error) {
			return s.repo.GetValuation(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, inventoryCacheTTL),
	)
}
