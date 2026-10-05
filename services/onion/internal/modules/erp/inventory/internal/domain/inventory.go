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
	ProductID int64
	Change    float64
	Type      InventoryTransactionType
	Notes     string
}

type ValuationItem struct {
	CategoryName  string
	WarehouseName string
	TotalItems    int64
	TotalValue    float64
	TotalCost     float64
}

type InventoryReportItem struct {
	ProductID         int64
	ProductName       string
	SKU               string
	Unit              string
	StockQuantity     float64
	LowStockThreshold float64
	ReorderQuantity   float64
	CostPrice         float64
	SalePrice         float64
	StockValue        float64
	IsLowStock        bool
}

type InventoryReportPage struct {
	Items      []InventoryReportItem
	Total      int64
	TotalValue float64
	HasMore    bool
	Limit      int64
	Offset     int64
}

type LowStockPage struct {
	Items   []InventoryReportItem
	Total   int64
	HasMore bool
	Limit   int64
	Offset  int64
}

type ValuationPage struct {
	Items      []ValuationItem
	Total      int64
	TotalValue float64
	HasMore    bool
	Limit      int64
	Offset     int64
}
