package api

import (
	"encoding/json"
	"strings"

	"github.com/voxire/onion/internal/modules/products/internal/domain"
	erpproductv1 "github.com/voxire/voxie/gen/go/erp/product/v1"
)

// productUpdateInput keeps JSON presence when the raw update body is
// available. A nil present map is the fallback for direct method callers that
// only have the generated scalar-valued request type; in that case non-zero
// values are treated as supplied so an omitted field cannot overwrite the row.
func productUpdateInput(body *erpproductv1.UpdateProductBody, present map[string]json.RawMessage) domain.ProductInput {
	if body == nil {
		return domain.ProductInput{PresentFields: make(map[string]bool)}
	}
	if present == nil {
		present = inferredProductUpdateFields(body)
	}

	fields := make(map[string]bool, len(present))
	for field, raw := range present {
		if len(raw) > 0 {
			fields[field] = true
		}
	}

	return domain.ProductInput{
		CategoryID:        body.GetCategoryId(),
		SupplierID:        body.GetSupplierId(),
		Name:              body.GetName(),
		SKU:               body.GetSku(),
		Barcode:           body.GetBarcode(),
		Description:       body.GetDescription(),
		Unit:              body.GetUnit(),
		CostPrice:         body.GetCostPrice(),
		SalePrice:         body.GetSalePrice(),
		StockQuantity:     body.GetStockQuantity(),
		StockQuantitySet:  fields[domain.ProductFieldStockQuantity],
		ImageURLSet:       fields[domain.ProductFieldImageURL],
		PresentFields:     fields,
		LowStockThreshold: body.GetLowStockThreshold(),
		ReorderQuantity:   body.GetReorderQuantity(),
		IsActive:          body.GetIsActive(),
		TrackInventory:    body.GetTrackInventory(),
		BrandID:           body.GetBrandId(),
		CompareAtPrice:    body.GetCompareAtPrice(),
		TaxRate:           body.GetTaxRate(),
		Currency:          body.GetCurrency(),
		Weight:            body.GetWeight(),
		WeightUnit:        body.GetWeightUnit(),
		Length:            body.GetLength(),
		Width:             body.GetWidth(),
		Height:            body.GetHeight(),
		DimensionUnit:     body.GetDimensionUnit(),
		CountryOfOrigin:   body.GetCountryOfOrigin(),
		HSCode:            body.GetHsCode(),
		AllowBackorder:    body.GetAllowBackorder(),
		LeadTimeDays:      int(body.GetLeadTimeDays()),
		ImageURL:          body.GetImageUrl(),
	}
}

func inferredProductUpdateFields(body *erpproductv1.UpdateProductBody) map[string]json.RawMessage {
	fields := make(map[string]json.RawMessage)
	mark := func(field string) { fields[field] = json.RawMessage("present") }

	if body.CategoryId != nil {
		mark(domain.ProductFieldCategoryID)
	}
	if body.SupplierId != nil {
		mark(domain.ProductFieldSupplierID)
	}
	if body.GetName() != "" {
		mark(domain.ProductFieldName)
	}
	if body.GetSku() != "" {
		mark(domain.ProductFieldSKU)
	}
	if body.GetBarcode() != "" {
		mark(domain.ProductFieldBarcode)
	}
	if body.GetDescription() != "" {
		mark(domain.ProductFieldDescription)
	}
	if body.GetUnit() != "" {
		mark(domain.ProductFieldUnit)
	}
	if body.GetCostPrice() != 0 {
		mark(domain.ProductFieldCostPrice)
	}
	if body.GetSalePrice() != 0 {
		mark(domain.ProductFieldSalePrice)
	}
	if body.GetStockQuantity() != 0 {
		mark(domain.ProductFieldStockQuantity)
	}
	if body.GetLowStockThreshold() != 0 {
		mark(domain.ProductFieldLowStockThreshold)
	}
	if body.GetReorderQuantity() != 0 {
		mark(domain.ProductFieldReorderQuantity)
	}
	if body.GetIsActive() {
		mark(domain.ProductFieldIsActive)
	}
	if body.GetTrackInventory() {
		mark(domain.ProductFieldTrackInventory)
	}
	if body.GetBrandId() != 0 {
		mark(domain.ProductFieldBrandID)
	}
	if body.GetCompareAtPrice() != 0 {
		mark(domain.ProductFieldCompareAtPrice)
	}
	if body.GetTaxRate() != 0 {
		mark(domain.ProductFieldTaxRate)
	}
	if body.GetCurrency() != "" {
		mark(domain.ProductFieldCurrency)
	}
	if body.GetWeight() != 0 {
		mark(domain.ProductFieldWeight)
	}
	if body.GetWeightUnit() != "" {
		mark(domain.ProductFieldWeightUnit)
	}
	if body.GetLength() != 0 {
		mark(domain.ProductFieldLength)
	}
	if body.GetWidth() != 0 {
		mark(domain.ProductFieldWidth)
	}
	if body.GetHeight() != 0 {
		mark(domain.ProductFieldHeight)
	}
	if body.GetDimensionUnit() != "" {
		mark(domain.ProductFieldDimensionUnit)
	}
	if body.GetCountryOfOrigin() != "" {
		mark(domain.ProductFieldCountryOfOrigin)
	}
	if body.GetHsCode() != "" {
		mark(domain.ProductFieldHSCode)
	}
	if body.GetAllowBackorder() {
		mark(domain.ProductFieldAllowBackorder)
	}
	if body.GetLeadTimeDays() != 0 {
		mark(domain.ProductFieldLeadTimeDays)
	}
	if strings.TrimSpace(body.GetImageUrl()) != "" {
		mark(domain.ProductFieldImageURL)
	}
	return fields
}
