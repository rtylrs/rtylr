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
)

type StockTransferStatus string

const (
	StockTransferStatusDraft     StockTransferStatus = "draft"
	StockTransferStatusInTransit StockTransferStatus = "in_transit"
	StockTransferStatusReceived  StockTransferStatus = "received"
	StockTransferStatusCancelled StockTransferStatus = "cancelled"
)

type StockTransfer struct {
	ID              int64               `json:"id"`
	OwnerID         int64               `json:"owner_id"`
	BusinessID      int64               `json:"business_id"`
	FromWarehouseID int64               `json:"from_warehouse_id"`
	ToWarehouseID   int64               `json:"to_warehouse_id"`
	Status          StockTransferStatus `json:"status"`
	Notes           string              `json:"notes"`
	Items           []StockTransferItem `json:"items"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}
type StockTransferItem struct {
	ID         int64     `json:"id"`
	TransferID int64     `json:"transfer_id"`
	ProductID  int64     `json:"product_id"`
	Quantity   float64   `json:"quantity"`
	CreatedAt  time.Time `json:"created_at"`
}
type StockTransferInput struct {
	FromWarehouseID int64                    `json:"from_warehouse_id"`
	ToWarehouseID   int64                    `json:"to_warehouse_id"`
	Notes           string                   `json:"notes"`
	Items           []StockTransferItemInput `json:"items"`
}

type StockTransferItemInput struct {
	ProductID int64   `json:"product_id"`
	Quantity  float64 `json:"quantity"`
}

type StockTransferStatusInput struct {
	ID     int64               `json:"id"`
	Status StockTransferStatus `json:"status"`
}

func ValidateStockTransferInput(input StockTransferInput) error {
	if input.FromWarehouseID <= 0 {
		return fmt.Errorf("from_warehouse_id must be positive")
	}
	if input.ToWarehouseID <= 0 {
		return fmt.Errorf("to_warehouse_id must be positive")
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
