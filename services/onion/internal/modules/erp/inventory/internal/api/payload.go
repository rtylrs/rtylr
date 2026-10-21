package api

import (
	"github.com/voxire/onion/internal/modules/erp/inventory/internal/domain"
	erpinventoryv1 "github.com/voxire/voxie/gen/go/erp/inventory/v1"
)

func newInventoryReportItemPayload(item domain.InventoryReportItem) *erpinventoryv1.InventoryReportItem {
	return &erpinventoryv1.InventoryReportItem{
		ProductId:         item.ProductID,
		ProductName:       item.ProductName,
		Sku:               item.SKU,
		Unit:              item.Unit,
		StockQuantity:     item.StockQuantity,
		LowStockThreshold: item.LowStockThreshold,
		ReorderQuantity:   item.ReorderQuantity,
		CostPrice:         item.CostPrice,
		SalePrice:         item.SalePrice,
		StockValue:        item.StockValue,
		IsLowStock:        item.IsLowStock,
	}
}

func newValuationItemPayload(item domain.ValuationItem) *erpinventoryv1.ValuationItem {
	return &erpinventoryv1.ValuationItem{
		CategoryName:  item.CategoryName,
		WarehouseName: item.WarehouseName,
		TotalItems:    item.TotalItems,
		TotalValue:    item.TotalValue,
		TotalCost:     item.TotalCost,
	}
}
