package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/erp/inventory/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	inventoryCacheTTL = 5 * time.Minute

	// Inventory contracts expose offset-based bounds rather than page numbers.
	// Keep the historical 100-row default, cap page size, and keep offsets within
	// the signed 32-bit range used by the generated SQL parameters.
	defaultInventoryLimit int64 = 100
	maxInventoryLimit     int64 = 500
	maxInventoryOffset    int64 = 1<<31 - 1
)

type repository interface {
	AdjustInventory(ctx context.Context, ownerID, productID int64, change float64, txType domain.InventoryTransactionType, notes string) error
	GetInventoryReportPage(ctx context.Context, ownerID int64, limit, offset int) (domain.InventoryReportPage, error)
	GetLowStockProductsPage(ctx context.Context, ownerID int64, limit, offset int) (domain.LowStockPage, error)
	GetValuationPage(ctx context.Context, ownerID int64, limit, offset int) (domain.ValuationPage, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{
		repo: repo,
	}
}

// Inventory views live under the product namespace because they are derived
// from product stock — wiping ProductPrefix covers every inventory view too.
func invalidateInventoryScope(ctx context.Context, ownerID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ProductPrefix(ownerID))
}

func (s *Service) AdjustInventory(ctx context.Context, ownerID int64, input domain.InventoryAdjustInput) error {
	if err := s.repo.AdjustInventory(ctx, ownerID, input.ProductID, input.Change, input.Type, input.Notes); err != nil {
		return err
	}

	invalidateInventoryScope(ctx, ownerID)

	lowStock, lowErr := s.repo.GetLowStockProductsPage(ctx, ownerID, int(defaultInventoryLimit), 0)
	if lowErr != nil {
		logger.Ctx(ctx).Error("Failed to check low stock after adjustment", logger.Error(lowErr))
	} else if len(lowStock.Items) > 0 {
		logger.Ctx(ctx).Info("Low stock detected after inventory adjustment",
			logger.Any("owner_id", ownerID),
			logger.Any("low_stock_count", lowStock.Total),
		)
	}

	return nil
}

func (s *Service) GetInventoryReport(ctx context.Context, ownerID int64) ([]domain.InventoryReportItem, error) {
	page, err := s.GetInventoryReportPage(ctx, ownerID, defaultInventoryLimit, 0)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (s *Service) GetInventoryReportPage(ctx context.Context, ownerID, limit, offset int64) (domain.InventoryReportPage, error) {
	limit, offset = normalizeInventoryBounds(limit, offset)
	key := fmt.Sprintf("%s|l=%d|o=%d", cacheKeys.InventoryReport(ownerID), limit, offset) + "|domain=v2"
	return cacher.Of[domain.InventoryReportPage]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.InventoryReportPage, error) {
			page, err := s.repo.GetInventoryReportPage(ctx, ownerID, int(limit), int(offset)) // #nosec G115 -- bounds are normalized above.
			if err != nil {
				return domain.InventoryReportPage{}, err
			}
			page.Limit = limit
			page.Offset = offset
			page.HasMore = int64(len(page.Items))+offset < page.Total
			return page, nil
		},
		cacher.WithExpiry(ctx, inventoryCacheTTL),
	)
}

func (s *Service) GetLowStockProducts(ctx context.Context, ownerID int64) ([]domain.InventoryReportItem, error) {
	page, err := s.GetLowStockProductsPage(ctx, ownerID, defaultInventoryLimit, 0)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (s *Service) GetLowStockProductsPage(ctx context.Context, ownerID, limit, offset int64) (domain.LowStockPage, error) {
	limit, offset = normalizeInventoryBounds(limit, offset)
	key := fmt.Sprintf("%s|l=%d|o=%d", cacheKeys.InventoryLowStock(ownerID), limit, offset) + "|domain=v2"
	return cacher.Of[domain.LowStockPage]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.LowStockPage, error) {
			page, err := s.repo.GetLowStockProductsPage(ctx, ownerID, int(limit), int(offset)) // #nosec G115 -- bounds are normalized above.
			if err != nil {
				return domain.LowStockPage{}, err
			}
			page.Limit = limit
			page.Offset = offset
			page.HasMore = int64(len(page.Items))+offset < page.Total
			return page, nil
		},
		cacher.WithExpiry(ctx, inventoryCacheTTL),
	)
}

func (s *Service) GetValuation(ctx context.Context, ownerID int64) ([]domain.ValuationItem, error) {
	page, err := s.GetValuationPage(ctx, ownerID, defaultInventoryLimit, 0)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (s *Service) GetValuationPage(ctx context.Context, ownerID, limit, offset int64) (domain.ValuationPage, error) {
	limit, offset = normalizeInventoryBounds(limit, offset)
	key := fmt.Sprintf("%s|l=%d|o=%d", cacheKeys.InventoryValuation(ownerID), limit, offset) + "|domain=v2"
	return cacher.Of[domain.ValuationPage]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.ValuationPage, error) {
			page, err := s.repo.GetValuationPage(ctx, ownerID, int(limit), int(offset)) // #nosec G115 -- bounds are normalized above.
			if err != nil {
				return domain.ValuationPage{}, err
			}
			page.Limit = limit
			page.Offset = offset
			page.HasMore = int64(len(page.Items))+offset < page.Total
			return page, nil
		},
		cacher.WithExpiry(ctx, inventoryCacheTTL),
	)
}

func normalizeInventoryBounds(limit, offset int64) (int64, int64) {
	if limit < 1 {
		limit = defaultInventoryLimit
	}
	if limit > maxInventoryLimit {
		limit = maxInventoryLimit
	}
	if offset < 0 {
		offset = 0
	}
	if offset > maxInventoryOffset {
		offset = maxInventoryOffset
	}
	return limit, offset
}
