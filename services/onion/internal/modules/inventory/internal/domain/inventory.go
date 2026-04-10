package domain

import (
	"errors"
)

var ErrInventoryTransactionFailed = errors.New("inventory transaction failed")

type InventoryTransactionType string

const (
	InventoryTypePurchase   InventoryTransactionType = "purchase"
	InventoryTypeSale       InventoryTransactionType = "sale"
	InventoryTypeAdjustment InventoryTransactionType = "adjustment"
	InventoryTypeReturn     InventoryTransactionType = "return"
	InventoryTypeWaste      InventoryTransactionType = "waste"
)

type InventoryAdjustInput struct {
	ProductID int64                    `json:"productId"`
	Change    float64                  `json:"change"`
	Type      InventoryTransactionType `json:"type"`
	Notes     string                   `json:"notes"`
}

type ValuationItem struct {
	CategoryName  string  `json:"categoryName"`
	WarehouseName string  `json:"warehouseName"`
	TotalItems    int64   `json:"totalItems"`
	TotalValue    float64 `json:"totalValue"`
	TotalCost     float64 `json:"totalCost"`
}

type InventoryReportItem struct {
	ProductID         int64   `json:"productId"`
	ProductName       string  `json:"productName"`
	SKU               string  `json:"sku"`
	Unit              string  `json:"unit"`
	StockQuantity     float64 `json:"stockQuantity"`
	LowStockThreshold float64 `json:"lowStockThreshold"`
	ReorderQuantity   float64 `json:"reorderQuantity"`
	CostPrice         float64 `json:"costPrice"`
	SalePrice         float64 `json:"salePrice"`
	StockValue        float64 `json:"stockValue"`
	IsLowStock        bool    `json:"isLowStock"`
}
