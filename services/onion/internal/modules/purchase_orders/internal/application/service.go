package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/purchase_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/purchase_orders/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/logger"
)

const poCacheTTL = 5 * time.Minute

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) GetByID(ctx context.Context, ownerID, id int64) (domain.PurchaseOrder, error) {
	return s.repo.GetByID(ctx, ownerID, id)
}

func (s *Service) ListPurchaseOrders(ctx context.Context, ownerID int64, status string) ([]domain.PurchaseOrder, error) {
	cacheKey := fmt.Sprintf("purchase_orders:%d:%s", ownerID, status)

	return cacher.Of[[]domain.PurchaseOrder]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.PurchaseOrder, error) {
		return s.repo.ListPurchaseOrders(ctx, ownerID, status)
	}, cacher.WithExpiry(ctx, poCacheTTL))
}

func (s *Service) CreatePurchaseOrder(ctx context.Context, ownerID int64, input domain.PurchaseOrderInput, createdBy int64) (int64, error) {
	id, err := s.repo.CreatePurchaseOrder(ctx, ownerID, input, createdBy)
	if err != nil {
		return 0, err
	}
	s.invalidateCache(ctx, ownerID)
	return id, nil
}

// ApprovePurchaseOrder transitions a purchase order to the `approved` status.
// Returns domain.ErrPurchaseOrderNotFound when the PO doesn't exist for the caller,
// or domain.ErrPurchaseOrderAlreadyApproved when the PO is already approved.
func (s *Service) ApprovePurchaseOrder(ctx context.Context, ownerID, poID int64) error {
	status, err := s.repo.GetPurchaseOrderStatus(ctx, ownerID, poID)
	if err != nil {
		return err
	}

	if status == string(domain.POStatusApproved) {
		return domain.ErrPurchaseOrderAlreadyApproved
	}

	if err := s.repo.ApprovePurchaseOrder(ctx, ownerID, poID); err != nil {
		return err
	}

	s.invalidateCache(ctx, ownerID)

	return nil
}

func (s *Service) UpdatePurchaseOrderStatus(ctx context.Context, ownerID, id int64, status string) error {
	err := s.repo.UpdatePurchaseOrderStatus(ctx, ownerID, id, status)
	if err != nil {
		return err
	}

	if status == string(domain.POStatusReceived) {
		items, itemErr := s.repo.GetPurchaseOrderItems(ctx, id)
		if itemErr != nil {
			logger.Ctx(ctx).Error("Failed to get purchase order items for inventory update", logger.Error(itemErr))
			return itemErr
		}

		for _, item := range items {
			adjustErr := s.repo.AdjustInventory(
				ctx,
				ownerID,
				item.ProductID,
				item.QuantityOrdered,
				domain.InventoryTypePurchase,
				fmt.Sprintf("auto-received from purchase order %d", id),
			)
			if adjustErr != nil {
				logger.Ctx(ctx).Error("Failed to adjust inventory for received PO item",
					logger.Error(adjustErr),
					logger.Any("product_id", item.ProductID),
				)
				return adjustErr
			}
		}
	}

	s.invalidateCache(ctx, ownerID)
	return nil
}

func (s *Service) invalidateCache(ctx context.Context, ownerID int64) {
	prefix := fmt.Sprintf("purchase_orders:%d:", ownerID)
	cacher.Of[[]domain.PurchaseOrder]().Invalidate(ctx, []string{prefix})
}
