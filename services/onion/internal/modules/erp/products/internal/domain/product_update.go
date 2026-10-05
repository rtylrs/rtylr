package domain

import "strings"

const (
	ProductFieldCategoryID        = "category_id"
	ProductFieldSupplierID        = "supplier_id"
	ProductFieldName              = "name"
	ProductFieldSKU               = "sku"
	ProductFieldBarcode           = "barcode"
	ProductFieldDescription       = "description"
	ProductFieldUnit              = "unit"
	ProductFieldCostPrice         = "cost_price"
	ProductFieldSalePrice         = "sale_price"
	ProductFieldStockQuantity     = "stock_quantity"
	ProductFieldLowStockThreshold = "low_stock_threshold"
	ProductFieldReorderQuantity   = "reorder_quantity"
	ProductFieldIsActive          = "is_active"
	ProductFieldTrackInventory    = "track_inventory"
	ProductFieldBrandID           = "brand_id"
	ProductFieldCompareAtPrice    = "compare_at_price"
	ProductFieldTaxRate           = "tax_rate"
	ProductFieldCurrency          = "currency"
	ProductFieldWeight            = "weight"
	ProductFieldWeightUnit        = "weight_unit"
	ProductFieldLength            = "length"
	ProductFieldWidth             = "width"
	ProductFieldHeight            = "height"
	ProductFieldDimensionUnit     = "dimension_unit"
	ProductFieldCountryOfOrigin   = "country_of_origin"
	ProductFieldHSCode            = "hs_code"
	ProductFieldAllowBackorder    = "allow_backorder"
	ProductFieldLeadTimeDays      = "lead_time_days"
	ProductFieldImageURL          = "image_url"
)

// FieldPresent reports whether an update explicitly carried a field. A nil
// presence map means the caller is using the legacy full-input shape.
func (i ProductInput) FieldPresent(field string) bool {
	return i.PresentFields == nil || i.PresentFields[field]
}

// MergeProductUpdate overlays a presence-aware update onto the current row.
// The repository merges against the locked current row before a full-row update.
func MergeProductUpdate(current Product, input ProductInput) ProductInput {
	if input.PresentFields == nil {
		// Legacy/internal callers historically treated a non-empty URL as an
		// image replacement. Keep that behavior while explicit clears use the
		// presence bit below.
		if strings.TrimSpace(input.ImageURL) != "" {
			input.ImageURLSet = true
		}
		return input
	}

	merged := input
	if input.FieldPresent(ProductFieldStockQuantity) {
		merged.StockQuantitySet = true
	}
	if input.FieldPresent(ProductFieldImageURL) {
		merged.ImageURLSet = true
	}
	if !input.FieldPresent(ProductFieldCategoryID) {
		merged.CategoryID = current.CategoryID
	}
	if !input.FieldPresent(ProductFieldSupplierID) {
		merged.SupplierID = current.SupplierID
	}
	if !input.FieldPresent(ProductFieldName) {
		merged.Name = current.Name
	}
	if !input.FieldPresent(ProductFieldSKU) {
		merged.SKU = current.SKU
	}
	if !input.FieldPresent(ProductFieldBarcode) {
		merged.Barcode = current.Barcode
	}
	if !input.FieldPresent(ProductFieldDescription) {
		merged.Description = current.Description
	}
	if !input.FieldPresent(ProductFieldUnit) {
		merged.Unit = current.Unit
	}
	if !input.FieldPresent(ProductFieldCostPrice) {
		merged.CostPrice = current.CostPrice
	}
	if !input.FieldPresent(ProductFieldSalePrice) {
		merged.SalePrice = current.SalePrice
	}
	if !merged.StockQuantitySet {
		merged.StockQuantity = current.StockQuantity
	}
	if !input.FieldPresent(ProductFieldLowStockThreshold) {
		merged.LowStockThreshold = current.LowStockThreshold
	}
	if !input.FieldPresent(ProductFieldReorderQuantity) {
		merged.ReorderQuantity = current.ReorderQuantity
	}
	if !input.FieldPresent(ProductFieldIsActive) {
		merged.IsActive = current.IsActive
	}
	if !input.FieldPresent(ProductFieldTrackInventory) {
		merged.TrackInventory = current.TrackInventory
	}
	if !input.FieldPresent(ProductFieldBrandID) {
		merged.BrandID = current.BrandID
	}
	if !input.FieldPresent(ProductFieldCompareAtPrice) {
		merged.CompareAtPrice = current.CompareAtPrice
	}
	if !input.FieldPresent(ProductFieldTaxRate) {
		merged.TaxRate = current.TaxRate
	}
	if !input.FieldPresent(ProductFieldCurrency) {
		merged.Currency = current.Currency
	}
	if !input.FieldPresent(ProductFieldWeight) {
		merged.Weight = current.Weight
	}
	if !input.FieldPresent(ProductFieldWeightUnit) {
		merged.WeightUnit = current.WeightUnit
	}
	if !input.FieldPresent(ProductFieldLength) {
		merged.Length = current.Length
	}
	if !input.FieldPresent(ProductFieldWidth) {
		merged.Width = current.Width
	}
	if !input.FieldPresent(ProductFieldHeight) {
		merged.Height = current.Height
	}
	if !input.FieldPresent(ProductFieldDimensionUnit) {
		merged.DimensionUnit = current.DimensionUnit
	}
	if !input.FieldPresent(ProductFieldCountryOfOrigin) {
		merged.CountryOfOrigin = current.CountryOfOrigin
	}
	if !input.FieldPresent(ProductFieldHSCode) {
		merged.HSCode = current.HSCode
	}
	if !input.FieldPresent(ProductFieldAllowBackorder) {
		merged.AllowBackorder = current.AllowBackorder
	}
	if !input.FieldPresent(ProductFieldLeadTimeDays) {
		merged.LeadTimeDays = current.LeadTimeDays
	}
	if !merged.ImageURLSet {
		merged.ImageURL = current.ImageURL
	}
	return merged
}
