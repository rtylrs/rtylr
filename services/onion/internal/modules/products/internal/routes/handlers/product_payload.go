package handlers

import (
	"time"

	"github.com/voxire/onion/internal/modules/products/internal/domain"
	erpproductv1 "github.com/voxire/voxie/gen/go/erp/product/v1"
)

func newProductPayload(product domain.Product) *erpproductv1.Product {
	return &erpproductv1.Product{
		Id:                product.ID,
		OwnerId:           product.OwnerID,
		CategoryId:        product.CategoryID,
		SupplierId:        product.SupplierID,
		Name:              product.Name,
		Sku:               product.SKU,
		Barcode:           product.Barcode,
		Description:       product.Description,
		Unit:              product.Unit,
		CostPrice:         product.CostPrice,
		SalePrice:         product.SalePrice,
		StockQuantity:     product.StockQuantity,
		LowStockThreshold: product.LowStockThreshold,
		ReorderQuantity:   product.ReorderQuantity,
		IsActive:          product.IsActive,
		TrackInventory:    product.TrackInventory,
		CreatedAt:         formatProductTimestamp(product.CreatedAt),
		UpdatedAt:         formatProductTimestamp(product.UpdatedAt),
	}
}

func formatProductTimestamp(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
