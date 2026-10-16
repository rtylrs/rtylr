package domain

import (
	"errors"
	"fmt"
	"math"
	"time"
)

var (
	ErrStockTransferNotFound          = errors.New("stock transfer not found")
	ErrStockTransferReferenceNotFound = errors.New("stock transfer reference not found in business")
	ErrInvalidStockTransfer           = errors.New("invalid stock transfer")
	ErrInvalidStockTransferTransition = errors.New("invalid stock transfer status transition")
)

type StockTransferStatus string

const (
	StockTransferStatusDraft     StockTransferStatus = "draft"
	StockTransferStatusInTransit StockTransferStatus = "in_transit"
	StockTransferStatusReceived  StockTransferStatus = "received"
	StockTransferStatusCancelled StockTransferStatus = "cancelled"
)

type StockTransfer struct {
	ID              int64
	OwnerID         int64
	BusinessID      int64
	FromWarehouseID int64
	ToWarehouseID   int64
	Status          StockTransferStatus
	Notes           string
	Items           []StockTransferItem
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type StockTransferItem struct {
	ID         int64
	TransferID int64
	ProductID  int64
	Quantity   float64
	CreatedAt  time.Time
}

type StockTransferInput struct {
	FromWarehouseID int64
	ToWarehouseID   int64
	Notes           string
	Items           []StockTransferItemInput
}

type StockTransferItemInput struct {
	ProductID int64
	Quantity  float64
}

type StockTransferStatusInput struct {
	ID     int64
	Status StockTransferStatus
}

func ValidateStockTransferInput(input StockTransferInput) error {
	if input.FromWarehouseID <= 0 {
		return fmt.Errorf("from_warehouse_id must be positive")
	}
	if input.ToWarehouseID <= 0 {
		return fmt.Errorf("to_warehouse_id must be positive")
	}
	if input.FromWarehouseID == input.ToWarehouseID {
		return fmt.Errorf("from_warehouse_id and to_warehouse_id must be different")
	}
	if len(input.Items) == 0 {
		return fmt.Errorf("at least one transfer item is required")
	}
	for i, item := range input.Items {
		if item.ProductID <= 0 {
			return fmt.Errorf("item[%d]: product_id must be positive", i)
		}
		if item.Quantity <= 0 || math.IsNaN(item.Quantity) || math.IsInf(item.Quantity, 0) {
			return fmt.Errorf("item[%d]: quantity must be a finite value greater than zero", i)
		}
	}
	return nil
}

func ValidateStockTransferStatusTransition(current, next StockTransferStatus) error {
	if current == next {
		return nil
	}
	if current == StockTransferStatusReceived || current == StockTransferStatusCancelled {
		return ErrInvalidStockTransferTransition
	}
	switch current {
	case StockTransferStatusDraft:
		if next == StockTransferStatusInTransit || next == StockTransferStatusReceived || next == StockTransferStatusCancelled {
			return nil
		}
	case StockTransferStatusInTransit:
		if next == StockTransferStatusReceived || next == StockTransferStatusCancelled {
			return nil
		}
	}
	return ErrInvalidStockTransferTransition
}
