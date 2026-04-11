package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/pos_stock_movements/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_stock_movements/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) ListMovements(ctx context.Context, businessID, branchID, productID int64) ([]domain.StockMovement, error) {
	return s.repo.ListMovements(ctx, businessID, branchID, productID)
}

func (s *Service) ReceiveStock(ctx context.Context, input domain.ReceiveStockInput) (int64, error) {
	id, err := s.repo.CreateMovement(ctx, input.BusinessID, input.BranchID, input.ProductID, "receive", input.Quantity, nil, nil, input.Note, input.EmployeeID)
	if err != nil {
		return 0, err
	}
	if err := s.repo.UpdateInventory(ctx, input.BusinessID, input.BranchID, input.ProductID, input.Quantity); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Service) AdjustStock(ctx context.Context, input domain.AdjustStockInput) (int64, error) {
	id, err := s.repo.CreateMovement(ctx, input.BusinessID, input.BranchID, input.ProductID, "adjust", input.Quantity, nil, nil, input.Note, input.EmployeeID)
	if err != nil {
		return 0, err
	}
	if err := s.repo.UpdateInventory(ctx, input.BusinessID, input.BranchID, input.ProductID, input.Quantity); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Service) RecordWaste(ctx context.Context, input domain.WasteStockInput) (int64, error) {
	id, err := s.repo.CreateMovement(ctx, input.BusinessID, input.BranchID, input.ProductID, "waste", -input.Quantity, nil, nil, input.Note, input.EmployeeID)
	if err != nil {
		return 0, err
	}
	if err := s.repo.UpdateInventory(ctx, input.BusinessID, input.BranchID, input.ProductID, -input.Quantity); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Service) TransferStock(ctx context.Context, input domain.TransferStockInput) (int64, error) {
	// Record transfer_out from source branch
	_, err := s.repo.CreateMovement(ctx, input.BusinessID, input.FromBranchID, input.ProductID, "transfer_out", -input.Quantity, nil, nil, input.Note, input.EmployeeID)
	if err != nil {
		return 0, err
	}
	if err := s.repo.UpdateInventory(ctx, input.BusinessID, input.FromBranchID, input.ProductID, -input.Quantity); err != nil {
		return 0, err
	}

	// Record transfer_in to destination branch
	id, err := s.repo.CreateMovement(ctx, input.BusinessID, input.ToBranchID, input.ProductID, "transfer_in", input.Quantity, nil, nil, input.Note, input.EmployeeID)
	if err != nil {
		return 0, err
	}
	if err := s.repo.UpdateInventory(ctx, input.BusinessID, input.ToBranchID, input.ProductID, input.Quantity); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Service) ListBranchInventory(ctx context.Context, businessID, branchID int64) ([]domain.BranchInventory, error) {
	return s.repo.ListBranchInventory(ctx, businessID, branchID)
}

func (s *Service) ListLowStockItems(ctx context.Context, businessID, branchID int64) ([]domain.BranchInventory, error) {
	return s.repo.ListLowStockItems(ctx, businessID, branchID)
}
