package products

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/erp/products/internal/domain"
)

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

func reorderSuggestionToolPayload(suggestion domain.ReorderSuggestion) map[string]any {
	return map[string]any{
		"product_id":         suggestion.ProductID,
		"name":               suggestion.Name,
		"sku":                suggestion.SKU,
		"current_stock":      suggestion.CurrentStock,
		"reorder_point":      suggestion.ReorderPoint,
		"suggested_quantity": suggestion.SuggestedQty,
		"avg_daily_sales":    suggestion.AvgDailySales,
		"lead_time_days":     suggestion.LeadTimeDays,
		"confidence":         suggestion.Confidence,
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
