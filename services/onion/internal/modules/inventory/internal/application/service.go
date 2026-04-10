package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/inventory/internal/domain"
	"github.com/voxire/onion/internal/modules/inventory/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/logger"
)

const inventoryCacheTTL = 5 * time.Minute

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	repo := infrastructure.NewRepository(ctx)

	return &Service{
		repo: repo,
	}
}

func (s *Service) AdjustInventory(ctx context.Context, ownerID int64, input domain.InventoryAdjustInput) error {
	err := s.repo.AdjustInventory(ctx, ownerID, input.ProductID, input.Change, input.Type, input.Notes)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, ownerID)

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
	cacheKey := fmt.Sprintf("inventory:report:%d", ownerID)

	return cacher.Of[[]domain.InventoryReportItem]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.InventoryReportItem, error) {
		return s.repo.GetInventoryReport(ctx, ownerID)
	}, cacher.WithExpiry(ctx, inventoryCacheTTL))
}

func (s *Service) GetLowStockProducts(ctx context.Context, ownerID int64) ([]domain.InventoryReportItem, error) {
	cacheKey := fmt.Sprintf("inventory:low_stock:%d", ownerID)

	return cacher.Of[[]domain.InventoryReportItem]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.InventoryReportItem, error) {
		return s.repo.GetLowStockProducts(ctx, ownerID)
	}, cacher.WithExpiry(ctx, inventoryCacheTTL))
}

func (s *Service) GetValuation(ctx context.Context, ownerID int64) ([]domain.ValuationItem, error) {
	cacheKey := fmt.Sprintf("inventory:valuation:%d", ownerID)

	return cacher.Of[[]domain.ValuationItem]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.ValuationItem, error) {
		return s.repo.GetValuation(ctx, ownerID)
	}, cacher.WithExpiry(ctx, inventoryCacheTTL))
}

func (s *Service) invalidateCache(ctx context.Context, ownerID int64) {
	reportKey := fmt.Sprintf("inventory:report:%d", ownerID)
	lowStockKey := fmt.Sprintf("inventory:low_stock:%d", ownerID)
	cacher.Of[[]domain.InventoryReportItem]().Invalidate(ctx, []string{reportKey, lowStockKey})
}
