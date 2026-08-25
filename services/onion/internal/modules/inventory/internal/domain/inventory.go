package domain

import (
	"errors"
)

var (
	ErrInventoryTransactionFailed = errors.New("inventory transaction failed")
	// ErrInsufficientStock is returned when an adjustment would drive stock
	// below zero. Negative stock is treated as data corruption, not a valid state.
	ErrInsufficientStock = errors.New("insufficient stock: adjustment would result in negative quantity")
)

type InventoryTransactionType string

const (
	InventoryTypePurchase   InventoryTransactionType = "purchase"
	InventoryTypeSale       InventoryTransactionType = "sale"
	InventoryTypeAdjustment InventoryTransactionType = "adjustment"
	InventoryTypeReturn     InventoryTransactionType = "return"
	InventoryTypeWaste      InventoryTransactionType = "waste"
)

type InventoryAdjustInput struct {
	ProductID int64                    `json:"product_id"`
	Change    float64                  `json:"change"`
	Type      InventoryTransactionType `json:"type"`
	Notes     string                   `json:"notes"`
}

type ValuationItem struct {
	CategoryName  string  `json:"category_name"`
	WarehouseName string  `json:"warehouse_name"`
	TotalItems    int64   `json:"total_items"`
	TotalValue    float64 `json:"total_value"`
	TotalCost     float64 `json:"total_cost"`
}

type InventoryReportItem struct {
	ProductID         int64   `json:"product_id"`
	ProductName       string  `json:"product_name"`
	SKU               string  `json:"sku"`
	Unit              string  `json:"unit"`
	StockQuantity     float64 `json:"stock_quantity"`
	LowStockThreshold float64 `json:"low_stock_threshold"`
	ReorderQuantity   float64 `json:"reorder_quantity"`
	CostPrice         float64 `json:"cost_price"`
	SalePrice         float64 `json:"sale_price"`
	StockValue        float64 `json:"stock_value"`
	IsLowStock        bool    `json:"is_low_stock"`
}

// CentsToFloat converts integer cents to standard float amount.
func CentsToFloat(cents int64) float64 {
	return float64(cents) / 100.0
}
