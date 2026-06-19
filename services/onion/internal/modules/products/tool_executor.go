package products

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/products/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	getProductToolName    = "erp.get_product"
	createProductToolName = "erp.create_product"
)
const maxProductToolLeadTimeDays = int64(1<<31 - 1)

type getProductToolInput struct {
	ProductID int64
	Barcode   string
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case getProductToolName:
		output, err := m.executeGetProductTool(ctx, input)
		return output, true, err
	case createProductToolName:
		output, err := m.executeCreateProductTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeGetProductTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseGetProductToolInput(input)
	if err != nil {
		return nil, err
	}

	ownerID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}

	var product domain.Product
	lookupBy := "id"
	if parsed.ProductID > 0 {
		product, err = m.service.GetProduct(ctx, ownerID, parsed.ProductID)
	} else {
		lookupBy = "barcode"
		product, err = m.service.GetProductByBarcode(ctx, ownerID, parsed.Barcode)
	}
	if err != nil {
		return nil, err
	}

	attrs, err := m.service.ListProductAttributes(ctx, ownerID, product.ID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"product":   productToolPayload(product, attrs),
		"lookup_by": lookupBy,
	}, nil
}

func (m *Module) executeCreateProductTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseCreateProductToolInput(input)
	if err != nil {
		return nil, err
	}

	ownerID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}
	id, err := m.service.CreateProduct(ctx, ownerID, parsed)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"product": productInputToolPayload(id, parsed),
	}, nil
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

func productToolNonNegativeInt64(value any) (int64, error) {
	out, err := productToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func productToolInt64(value any) (int64, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case float64:
		if math.Trunc(v) != v {
			return 0, errors.New("must be an integer")
		}
		return int64(v), nil
	case json.Number:
		return v.Int64()
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return 0, nil
		}
		return strconv.ParseInt(raw, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported value %T", value)
	}
}

func productToolNonNegativeFloat64(value any) (float64, error) {
	out, err := productToolFloat64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func productToolFloat64(value any) (float64, error) {
	var out float64
	switch v := value.(type) {
	case nil:
		return 0, nil
	case float64:
		out = v
	case float32:
		out = float64(v)
	case int:
		out = float64(v)
	case int64:
		out = float64(v)
	case int32:
		out = float64(v)
	case json.Number:
		parsed, err := v.Float64()
		if err != nil {
			return 0, err
		}
		out = parsed
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return 0, nil
		}
		parsed, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return 0, err
		}
		out = parsed
	default:
		return 0, fmt.Errorf("unsupported value %T", value)
	}
	if math.IsNaN(out) || math.IsInf(out, 0) {
		return 0, errors.New("must be finite")
	}
	return out, nil
}

func productToolBool(value any, defaultValue bool) (bool, error) {
	switch v := value.(type) {
	case nil:
		return defaultValue, nil
	case bool:
		return v, nil
	case json.Number:
		n, err := strconv.ParseInt(v.String(), 10, 64)
		if err != nil {
			return false, err
		}
		switch n {
		case 0:
			return false, nil
		case 1:
			return true, nil
		default:
			return false, errors.New("must be true or false")
		}
	case string:
		raw := strings.ToLower(strings.TrimSpace(v))
		if raw == "" {
			return defaultValue, nil
		}
		switch raw {
		case "true", "t", "yes", "y", "1", "active", "enabled", "on":
			return true, nil
		case "false", "f", "no", "n", "0", "inactive", "disabled", "off":
			return false, nil
		default:
			return false, errors.New("must be true or false")
		}
	default:
		return false, fmt.Errorf("unsupported value %T", value)
	}
}

func productToolString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case json.Number:
		return v.String()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}

func productInputToolPayload(id int64, product domain.ProductInput) map[string]any {
	return map[string]any{
		"id":                  id,
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
		"compare_at_price":    product.CompareAtPrice,
		"tax_rate":            product.TaxRate,
		"currency":            product.Currency,
		"stock_quantity":      product.StockQuantity,
		"low_stock_threshold": product.LowStockThreshold,
		"reorder_quantity":    product.ReorderQuantity,
		"is_active":           product.IsActive,
		"track_inventory":     product.TrackInventory,
		"allow_backorder":     product.AllowBackorder,
		"lead_time_days":      product.LeadTimeDays,
	}
}

func productToolPayload(product domain.Product, attrs []domain.ProductAttribute) map[string]any {
	attrPayload := make([]map[string]any, 0, len(attrs))
	for _, attr := range attrs {
		attrPayload = append(attrPayload, map[string]any{
			"id":       attr.ID,
			"name":     attr.Name,
			"value":    attr.Value,
			"position": attr.Position,
		})
	}

	return map[string]any{
		"id":                  product.ID,
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
		"compare_at_price":    product.CompareAtPrice,
		"gross_profit":        product.GrossProfit(),
		"margin_percent":      product.Margin(),
		"tax_rate":            product.TaxRate,
		"currency":            product.Currency,
		"stock_quantity":      product.StockQuantity,
		"stock_value":         product.StockValue(),
		"low_stock_threshold": product.LowStockThreshold,
		"reorder_quantity":    product.ReorderQuantity,
		"is_low_stock":        product.IsLowStock(),
		"is_active":           product.IsActive,
		"status":              string(product.Status()),
		"track_inventory":     product.TrackInventory,
		"allow_backorder":     product.AllowBackorder,
		"lead_time_days":      product.LeadTimeDays,
		"image_url":           product.ImageURL,
		"weight":              product.Weight,
		"weight_unit":         product.WeightUnit,
		"length":              product.Length,
		"width":               product.Width,
		"height":              product.Height,
		"dimension_unit":      product.DimensionUnit,
		"country_of_origin":   product.CountryOfOrigin,
		"hs_code":             product.HSCode,
		"attributes":          attrPayload,
	}
}

func validProductToolCurrency(currency string) bool {
	if len(currency) != 3 {
		return false
	}
	for _, r := range currency {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
