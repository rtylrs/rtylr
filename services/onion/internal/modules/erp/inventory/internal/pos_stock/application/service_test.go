package application

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/inventory/internal/pos_stock/domain"
)

// Service with nil repo panics after validation guard, so these tests only
// verify that the guards fire correctly.

func TestReceiveStock_RejectsMissingBusinessID(t *testing.T) {
	s := &Service{}
	_, err := s.ReceiveStock(context.Background(), domain.ReceiveStockInput{
		BranchID: 1, ProductID: 2, Quantity: 10,
	})
	if err == nil {
		t.Fatal("expected error for missing business id")
	}
}

func TestReceiveStock_RejectsMissingProductID(t *testing.T) {
	s := &Service{}
	_, err := s.ReceiveStock(context.Background(), domain.ReceiveStockInput{
		BusinessID: 1, OwnerID: 1, BranchID: 2, Quantity: 10,
	})
	if err == nil {
		t.Fatal("expected error for missing product id")
	}
}

func TestReceiveStock_RejectsZeroQuantity(t *testing.T) {
	s := &Service{}
	_, err := s.ReceiveStock(context.Background(), domain.ReceiveStockInput{
		BusinessID: 1, OwnerID: 1, BranchID: 2, ProductID: 3, Quantity: 0,
	})
	if err == nil {
		t.Fatal("expected error for zero quantity")
	}
}

func TestTransferStock_RejectsSameBranch(t *testing.T) {
	s := &Service{}
	_, err := s.TransferStock(context.Background(), domain.TransferStockInput{
		BusinessID: 1, OwnerID: 1, FromBranchID: 2, ToBranchID: 2, ProductID: 3, Quantity: 5,
	})
	if err == nil {
		t.Fatal("expected error for same source and destination branch")
	}
}

func TestTransferStock_RejectsZeroQuantity(t *testing.T) {
	s := &Service{}
	_, err := s.TransferStock(context.Background(), domain.TransferStockInput{
		BusinessID: 1, OwnerID: 1, FromBranchID: 2, ToBranchID: 3, ProductID: 4, Quantity: 0,
	})
	if err == nil {
		t.Fatal("expected error for zero quantity")
	}
}
