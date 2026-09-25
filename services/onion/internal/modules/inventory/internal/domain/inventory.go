package domain

import (
	"errors"
)

var (
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

type InventoryReportPage struct {
	Items      []InventoryReportItem `json:"items"`
	Total      int64                 `json:"total"`
	TotalValue float64               `json:"total_value"`
	HasMore    bool                  `json:"has_more"`
	Limit      int64                 `json:"limit"`
	Offset     int64                 `json:"offset"`
}

type LowStockPage struct {
	Items   []InventoryReportItem `json:"items"`
	Total   int64                 `json:"total"`
	HasMore bool                  `json:"has_more"`
	Limit   int64                 `json:"limit"`
	Offset  int64                 `json:"offset"`
}

type ValuationPage struct {
	Items      []ValuationItem `json:"items"`
	Total      int64           `json:"total"`
	TotalValue float64         `json:"total_value"`
	HasMore    bool            `json:"has_more"`
	Limit      int64           `json:"limit"`
	Offset     int64           `json:"offset"`
}
