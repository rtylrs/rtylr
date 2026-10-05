package application

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/voxire/onion/internal/modules/erp/inventory/internal/pos_stock/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

// Stock movements derive from product stock, so cache under ProductPrefix
// — any inventory adjustment from elsewhere wipes these views too.
const stockMovementsTTL = 1 * time.Minute

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func invalidateStockMovementsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ProductPrefix(businessID))
}

func (s *Service) ListMovements(ctx context.Context, businessID, ownerID, branchID, productID int64) ([]domain.StockMovement, error) {
	key := fmt.Sprintf("%s|stock|b=%d|p=%d"+"|domain=v2",
		cacheKeys.ProductsList(businessID), branchID, productID)
	return cacher.Of[[]domain.StockMovement]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.StockMovement, error) {
			return s.repo.ListMovements(ctx, businessID, ownerID, branchID, productID)
		},
		cacher.WithExpiry(ctx, stockMovementsTTL),
	)
}

func (s *Service) ReceiveStock(ctx context.Context, input domain.ReceiveStockInput) (int64, error) {
	if err := domain.ValidateReceiveStockInput(input); err != nil {
		return 0, err
	}
	id, err := s.repo.ReceiveStock(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateStockMovementsScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) AdjustStock(ctx context.Context, input domain.AdjustStockInput) (int64, error) {
	if err := domain.ValidateAdjustStockInput(input); err != nil {
		return 0, err
	}
	id, err := s.repo.AdjustStock(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateStockMovementsScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) RecordWaste(ctx context.Context, input domain.WasteStockInput) (int64, error) {
	if err := domain.ValidateWasteStockInput(input); err != nil {
		return 0, err
	}
	id, err := s.repo.RecordWaste(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateStockMovementsScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) TransferStock(ctx context.Context, input domain.TransferStockInput) (int64, error) {
	if err := domain.ValidateTransferStockInput(input); err != nil {
		return 0, err
	}
	id, err := s.repo.TransferStock(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateStockMovementsScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) ApplyRecordedOrderInventory(ctx context.Context, input domain.RecordedOrderInventory) error {
	if input.BusinessID <= 0 || input.OwnerID <= 0 || input.SourceID <= 0 || input.LocationKind == "" || input.Channel == "" {
		return errors.New("recorded order inventory is missing required identifiers")
	}
	if input.SourceType != "pos_order" && input.SourceType != "store_order" {
		return errors.New("recorded order inventory has an unsupported source type")
	}
	if input.Channel != "pos" && input.Channel != "erp" {
		return errors.New("recorded order inventory has an unsupported channel")
	}
	if len(input.Items) == 0 {
		return nil
	}
	for _, item := range input.Items {
		if item.ProductID <= 0 || item.Quantity <= 0 || math.IsNaN(item.Quantity) || math.IsInf(item.Quantity, 0) {
			return errors.New("recorded order inventory contains an invalid line item")
		}
	}
	if err := s.repo.ApplyRecordedOrderInventory(ctx, input); err != nil {
		return err
	}
	invalidateStockMovementsScope(ctx, input.BusinessID)
	return nil
}

func (s *Service) ListBranchInventory(ctx context.Context, businessID, ownerID, branchID int64) ([]domain.BranchInventory, error) {
	key := fmt.Sprintf("%s|branch_inv|b=%d", cacheKeys.ProductsList(businessID), branchID) + "|domain=v2"
	return cacher.Of[[]domain.BranchInventory]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.BranchInventory, error) {
			return s.repo.ListBranchInventory(ctx, businessID, ownerID, branchID)
		},
		cacher.WithExpiry(ctx, stockMovementsTTL),
	)
}

func (s *Service) ListLowStockItems(ctx context.Context, businessID, ownerID, branchID int64) ([]domain.BranchInventory, error) {
	key := fmt.Sprintf("%s|branch_low|b=%d", cacheKeys.ProductsList(businessID), branchID) + "|domain=v2"
	return cacher.Of[[]domain.BranchInventory]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.BranchInventory, error) {
			return s.repo.ListLowStockItems(ctx, businessID, ownerID, branchID)
		},
		cacher.WithExpiry(ctx, stockMovementsTTL),
	)
}

type repository interface {
	ApplyRecordedOrderInventory(ctx context.Context, input domain.RecordedOrderInventory) error
	AdjustStock(ctx context.Context, input domain.AdjustStockInput) (int64, error)
	ListBranchInventory(ctx context.Context, businessID, _, branchID int64) ([]domain.BranchInventory, error)
	ListLowStockItems(ctx context.Context, businessID, _, branchID int64) ([]domain.BranchInventory, error)
	ListMovements(ctx context.Context, businessID, _, branchID, productID int64) ([]domain.StockMovement, error)
	ReceiveStock(ctx context.Context, input domain.ReceiveStockInput) (int64, error)
	RecordWaste(ctx context.Context, input domain.WasteStockInput) (int64, error)
	TransferStock(ctx context.Context, input domain.TransferStockInput) (int64, error)
}
