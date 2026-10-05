package products

import (
	"errors"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/erp/products/internal/domain"
)

func parseProductToolID(input map[string]any) (int64, error) {
	productID, err := productToolNonNegativeInt64(firstProductToolValue(input, "product_id", "id"))
	if err != nil {
		return 0, fmt.Errorf("invalid product_id: %w", err)
	}
	if productID <= 0 {
		return 0, errors.New("product_id is required")
	}
	return productID, nil
}

var productUpdateToolKeys = map[string]struct{}{
	"name": {}, "product_name": {}, "sku": {}, "barcode": {}, "description": {}, "unit": {},
	"category_id": {}, "supplier_id": {}, "brand_id": {}, "cost_price": {}, "cost": {},
	"sale_price": {}, "price": {}, "unit_price": {}, "stock_quantity": {}, "quantity": {}, "stock": {},
	"low_stock_threshold": {}, "low_stock": {}, "minimum_stock": {}, "reorder_quantity": {}, "reorder_qty": {},
	"is_active": {}, "track_inventory": {}, "allow_backorder": {}, "lead_time_days": {}, "lead_time": {},
	"compare_at_price": {}, "compare_price": {}, "tax_rate": {}, "currency": {}, "weight": {},
	"weight_unit": {}, "length": {}, "width": {}, "height": {}, "dimension_unit": {},
	"country_of_origin": {}, "hs_code": {},
}

func mergeProductUpdateToolInput(product domain.Product, input map[string]any) (domain.ProductInput, error) {
	merged := map[string]any{
		"name":                product.Name,
		"sku":                 product.SKU,
		"barcode":             product.Barcode,
		"description":         product.Description,
		"unit":                product.Unit,
		"category_id":         product.CategoryID,
		"supplier_id":         product.SupplierID,
		"brand_id":            product.BrandID,
		"cost_price":          product.CostPrice,
		"sale_price":          product.SalePrice,
		"stock_quantity":      product.StockQuantity,
		"low_stock_threshold": product.LowStockThreshold,
		"reorder_quantity":    product.ReorderQuantity,
		"is_active":           product.IsActive,
		"track_inventory":     product.TrackInventory,
		"allow_backorder":     product.AllowBackorder,
		"lead_time_days":      product.LeadTimeDays,
		"compare_at_price":    product.CompareAtPrice,
		"tax_rate":            product.TaxRate,
		"currency":            product.Currency,
		"weight":              product.Weight,
		"weight_unit":         product.WeightUnit,
		"length":              product.Length,
		"width":               product.Width,
		"height":              product.Height,
		"dimension_unit":      product.DimensionUnit,
		"country_of_origin":   product.CountryOfOrigin,
		"hs_code":             product.HSCode,
	}
	changed := false
	for key, value := range input {
		if key == "id" || key == "product_id" {
			continue
		}
		if _, ok := productUpdateToolKeys[key]; !ok {
			return domain.ProductInput{}, fmt.Errorf("unsupported product field %q", key)
		}
		merged[canonicalProductUpdateToolKey(key)] = value
		changed = true
	}
	if !changed {
		return domain.ProductInput{}, errors.New("at least one product field to update is required")
	}

	parsed, err := parseCreateProductToolInput(merged)
	if err != nil {
		return domain.ProductInput{}, err
	}
	parsed.StockQuantitySet = productToolHasAny(input, "stock_quantity", "quantity", "stock")
	return parsed, nil
}

func canonicalProductUpdateToolKey(key string) string {
	switch key {
	case "product_name":
		return "name"
	case "cost":
		return "cost_price"
	case "price", "unit_price":
		return "sale_price"
	case "quantity", "stock":
		return "stock_quantity"
	case "low_stock", "minimum_stock":
		return "low_stock_threshold"
	case "reorder_qty":
		return "reorder_quantity"
	case "lead_time":
		return "lead_time_days"
	case "compare_price":
		return "compare_at_price"
	default:
		return key
	}
}

func productToolHasAny(input map[string]any, keys ...string) bool {
	for _, key := range keys {
		if value, ok := input[key]; ok && value != nil {
			return true
		}
	}
	return false
}

func parseGetProductToolInput(input map[string]any) (getProductToolInput, error) {
	productID, err := productToolNonNegativeInt64(firstProductToolValue(input, "product_id", "id"))
	if err != nil {
		return getProductToolInput{}, fmt.Errorf("invalid product_id: %w", err)
	}
	barcode := strings.TrimSpace(productToolString(firstProductToolValue(input, "barcode", "bar_code")))
	if productID == 0 && barcode == "" {
		return getProductToolInput{}, errors.New("product_id or barcode is required")
	}
	return getProductToolInput{
		ProductID: productID,
		Barcode:   barcode,
	}, nil
}

func parseCreateProductToolInput(input map[string]any) (domain.ProductInput, error) {
	name := strings.TrimSpace(productToolString(firstProductToolValue(input, "name", "product_name")))
	if name == "" {
		return domain.ProductInput{}, errors.New("name is required")
	}

	categoryID, err := productToolNonNegativeInt64(firstProductToolValue(input, "category_id"))
	if err != nil {
		return domain.ProductInput{}, fmt.Errorf("invalid category_id: %w", err)
	}
	supplierID, err := productToolNonNegativeInt64(firstProductToolValue(input, "supplier_id"))
	if err != nil {
		return domain.ProductInput{}, fmt.Errorf("invalid supplier_id: %w", err)
	}
	brandID, err := productToolNonNegativeInt64(firstProductToolValue(input, "brand_id"))
	if err != nil {
		return domain.ProductInput{}, fmt.Errorf("invalid brand_id: %w", err)
	}

	costPrice, err := productToolNonNegativeFloat64(firstProductToolValue(input, "cost_price", "cost"))
	if err != nil {
		return domain.ProductInput{}, fmt.Errorf("invalid cost_price: %w", err)
	}
	salePrice, err := productToolNonNegativeFloat64(firstProductToolValue(input, "sale_price", "price", "unit_price"))
	if err != nil {
		return domain.ProductInput{}, fmt.Errorf("invalid sale_price: %w", err)
	}
	stockQuantity, err := productToolNonNegativeFloat64(firstProductToolValue(input, "stock_quantity", "quantity", "stock"))
	if err != nil {
		return domain.ProductInput{}, fmt.Errorf("invalid stock_quantity: %w", err)
	}
	lowStockThreshold, err := productToolNonNegativeFloat64(firstProductToolValue(input, "low_stock_threshold", "low_stock", "minimum_stock"))
	if err != nil {
		return domain.ProductInput{}, fmt.Errorf("invalid low_stock_threshold: %w", err)
	}
	reorderQuantity, err := productToolNonNegativeFloat64(firstProductToolValue(input, "reorder_quantity", "reorder_qty"))
	if err != nil {
		return domain.ProductInput{}, fmt.Errorf("invalid reorder_quantity: %w", err)
	}
	compareAtPrice, err := productToolNonNegativeFloat64(firstProductToolValue(input, "compare_at_price", "compare_price"))
	if err != nil {
		return domain.ProductInput{}, fmt.Errorf("invalid compare_at_price: %w", err)
	}
	taxRate, err := productToolNonNegativeFloat64(firstProductToolValue(input, "tax_rate"))
	if err != nil {
		return domain.ProductInput{}, fmt.Errorf("invalid tax_rate: %w", err)
	}
	if taxRate > 100 {
		return domain.ProductInput{}, errors.New("tax_rate must be between 0 and 100")
	}

	weight, err := productToolNonNegativeFloat64(firstProductToolValue(input, "weight"))
	if err != nil {
		return domain.ProductInput{}, fmt.Errorf("invalid weight: %w", err)
	}
	length, err := productToolNonNegativeFloat64(firstProductToolValue(input, "length"))
	if err != nil {
		return domain.ProductInput{}, fmt.Errorf("invalid length: %w", err)
	}
	width, err := productToolNonNegativeFloat64(firstProductToolValue(input, "width"))
	if err != nil {
		return domain.ProductInput{}, fmt.Errorf("invalid width: %w", err)
	}
	height, err := productToolNonNegativeFloat64(firstProductToolValue(input, "height"))
	if err != nil {
		return domain.ProductInput{}, fmt.Errorf("invalid height: %w", err)
	}

	isActive, err := productToolBool(firstProductToolValue(input, "is_active"), true)
	if err != nil {
		return domain.ProductInput{}, fmt.Errorf("invalid is_active: %w", err)
	}
	trackInventory, err := productToolBool(firstProductToolValue(input, "track_inventory"), true)
	if err != nil {
		return domain.ProductInput{}, fmt.Errorf("invalid track_inventory: %w", err)
	}
	allowBackorder, err := productToolBool(firstProductToolValue(input, "allow_backorder"), false)
	if err != nil {
		return domain.ProductInput{}, fmt.Errorf("invalid allow_backorder: %w", err)
	}

	leadTimeDays64, err := productToolNonNegativeInt64(firstProductToolValue(input, "lead_time_days", "lead_time"))
	if err != nil {
		return domain.ProductInput{}, fmt.Errorf("invalid lead_time_days: %w", err)
	}
	if leadTimeDays64 > maxProductToolLeadTimeDays {
		return domain.ProductInput{}, errors.New("lead_time_days is too large")
	}

	currency := strings.ToUpper(strings.TrimSpace(productToolString(firstProductToolValue(input, "currency"))))
	if currency != "" && !validProductToolCurrency(currency) {
		return domain.ProductInput{}, errors.New("currency must be a 3-letter code")
	}

	parsed := domain.ProductInput{
		CategoryID:        categoryID,
		SupplierID:        supplierID,
		Name:              name,
		SKU:               strings.TrimSpace(productToolString(firstProductToolValue(input, "sku"))),
		Barcode:           strings.TrimSpace(productToolString(firstProductToolValue(input, "barcode"))),
		Description:       strings.TrimSpace(productToolString(firstProductToolValue(input, "description"))),
		Unit:              strings.TrimSpace(productToolString(firstProductToolValue(input, "unit"))),
		CostPrice:         costPrice,
		SalePrice:         salePrice,
		StockQuantity:     stockQuantity,
		StockQuantitySet:  true,
		LowStockThreshold: lowStockThreshold,
		ReorderQuantity:   reorderQuantity,
		IsActive:          isActive,
		TrackInventory:    trackInventory,
		BrandID:           brandID,
		CompareAtPrice:    compareAtPrice,
		TaxRate:           taxRate,
		Currency:          currency,
		Weight:            weight,
		WeightUnit:        strings.TrimSpace(productToolString(firstProductToolValue(input, "weight_unit"))),
		Length:            length,
		Width:             width,
		Height:            height,
		DimensionUnit:     strings.TrimSpace(productToolString(firstProductToolValue(input, "dimension_unit"))),
		CountryOfOrigin:   strings.TrimSpace(productToolString(firstProductToolValue(input, "country_of_origin"))),
		HSCode:            strings.TrimSpace(productToolString(firstProductToolValue(input, "hs_code"))),
		AllowBackorder:    allowBackorder,
		LeadTimeDays:      int(leadTimeDays64),
		ImageURL:          strings.TrimSpace(productToolString(firstProductToolValue(input, "image_url", "image"))),
	}
	if err := parsed.Validate(); err != nil {
		return domain.ProductInput{}, err
	}
	return parsed, nil
}

func firstProductToolValue(input map[string]any, keys ...string) any {
	for _, key := range keys {
		value, ok := input[key]
		if !ok || value == nil {
			continue
		}
		if s, ok := value.(string); ok && strings.TrimSpace(s) == "" {
			continue
		}
		return value
	}
	return nil
}
