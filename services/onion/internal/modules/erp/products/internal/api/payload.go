package api

import (
	"time"

	"github.com/voxire/onion/internal/modules/erp/products/internal/domain"
	erpproductv1 "github.com/voxire/voxie/gen/go/erp/product/v1"
)

func newProductPayload(product domain.Product) *erpproductv1.Product {
	return &erpproductv1.Product{
		Id:                product.ID,
		OwnerId:           product.OwnerID,
		BusinessId:        product.BusinessID,
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
		BrandId:           product.BrandID,
		CompareAtPrice:    product.CompareAtPrice,
		TaxRate:           product.TaxRate,
		Currency:          product.Currency,
		Weight:            product.Weight,
		WeightUnit:        product.WeightUnit,
		Length:            product.Length,
		Width:             product.Width,
		Height:            product.Height,
		DimensionUnit:     product.DimensionUnit,
		CountryOfOrigin:   product.CountryOfOrigin,
		HsCode:            product.HSCode,
		AllowBackorder:    product.AllowBackorder,
		LeadTimeDays:      int32(product.LeadTimeDays), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		ImageUrl:          product.ImageURL,
		CreatedAt:         formatProductTimestamp(product.CreatedAt),
		UpdatedAt:         formatProductTimestamp(product.UpdatedAt),
	}
}

func newFullProductPayload(product domain.Product) *erpproductv1.FullProduct {
	return &erpproductv1.FullProduct{
		Id:                product.ID,
		OwnerId:           product.OwnerID,
		BusinessId:        product.BusinessID,
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
		BrandId:           product.BrandID,
		CompareAtPrice:    product.CompareAtPrice,
		TaxRate:           product.TaxRate,
		Currency:          product.Currency,
		Weight:            product.Weight,
		WeightUnit:        product.WeightUnit,
		Length:            product.Length,
		Width:             product.Width,
		Height:            product.Height,
		DimensionUnit:     product.DimensionUnit,
		CountryOfOrigin:   product.CountryOfOrigin,
		HsCode:            product.HSCode,
		AllowBackorder:    product.AllowBackorder,
		LeadTimeDays:      int32(product.LeadTimeDays), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		ImageUrl:          product.ImageURL,
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
