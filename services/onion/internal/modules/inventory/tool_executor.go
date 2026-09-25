package inventory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/inventory/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

type setStockToolInput struct {
	ProductID int64
	Quantity  float64
	Notes     string
}

const (
	listProductsToolName         = "erp.list_products"
	setStockToolName             = "erp.set_stock"
	defaultListProductsToolLimit = 20
	maxListProductsToolPageSize  = 50
)

type listProductsToolInput struct {
	Search       string
	LowStockOnly bool
	Page         int
	PageSize     int
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listProductsToolName:
		output, err := m.executeListProductsTool(ctx, input)
		return output, true, err
	case setStockToolName:
		output, err := m.executeSetStockTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListProductsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseListProductsToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}
	report, err := m.service.GetInventoryReport(ctx, businessID)
	if err != nil {
		return nil, err
	}

	filtered := filterInventoryReportItems(report, parsed)
	pageItems := paginateInventoryReportItems(filtered, parsed.Page, parsed.PageSize)
	products := make([]map[string]any, 0, len(pageItems))
	for _, item := range pageItems {
		products = append(products, inventoryReportToolPayload(item))
	}

	return map[string]any{
		"products":          products,
		"total":             len(filtered),
		"inventory_total":   len(report),
		"returned":          len(products),
		"page":              parsed.Page,
		"page_size":         parsed.PageSize,
		"low_stock_count":   domain.LowStockCount(report),
		"total_stock_value": domain.TotalStockValue(report),
		"filters": map[string]any{
			"search":         parsed.Search,
			"low_stock_only": parsed.LowStockOnly,
		},
	}, nil
}

func (m *Module) executeSetStockTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseSetStockToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}
	report, err := m.service.GetInventoryReport(ctx, businessID)
	if err != nil {
		return nil, err
	}

	var product *domain.InventoryReportItem
	for i := range report {
		if report[i].ProductID == parsed.ProductID {
			product = &report[i]
			break
		}
	}
	if product == nil {
		return nil, errors.New("product was not found in the tenant inventory report")
	}

	current := product.StockQuantity
	change := parsed.Quantity - current
	if change != 0 {
		notes := parsed.Notes
		if notes == "" {
			notes = fmt.Sprintf("Caret set stock from %s to %s", formatStockQuantity(current), formatStockQuantity(parsed.Quantity))
		}
		if err := m.service.AdjustInventory(ctx, businessID, domain.InventoryAdjustInput{
			ProductID: parsed.ProductID,
			Change:    change,
			Type:      domain.InventoryTypeAdjustment,
			Notes:     notes,
		}); err != nil {
			return nil, err
		}
	}

	return map[string]any{
		"product": map[string]any{
			"id":                   parsed.ProductID,
			"name":                 product.ProductName,
			"sku":                  product.SKU,
			"previous_quantity":    current,
			"stock_quantity":       parsed.Quantity,
			"quantity_change":      change,
			"low_stock_threshold":  product.LowStockThreshold,
			"would_be_low_stock":   parsed.Quantity <= product.LowStockThreshold,
			"inventory_adjustment": change != 0,
		},
	}, nil
}

func parseSetStockToolInput(input map[string]any) (setStockToolInput, error) {
	productID, err := stockToolInt64(input["product_id"])
	if err != nil {
		return setStockToolInput{}, fmt.Errorf("invalid product_id: %w", err)
	}
	if productID <= 0 {
		return setStockToolInput{}, errors.New("product_id is required")
	}

	quantityValue, ok := input["quantity"]
	if !ok {
		quantityValue, ok = input["stock_quantity"]
	}
	if !ok {
		return setStockToolInput{}, errors.New("quantity is required")
	}
	if quantityValue == nil {
		return setStockToolInput{}, errors.New("quantity is required")
	}
	quantity, err := stockToolFloat64(quantityValue)
	if err != nil {
		return setStockToolInput{}, fmt.Errorf("invalid quantity: %w", err)
	}
	if quantity < 0 {
		return setStockToolInput{}, errors.New("quantity must be non-negative")
	}

	return setStockToolInput{
		ProductID: productID,
		Quantity:  quantity,
		Notes:     strings.TrimSpace(stockToolString(input["notes"])),
	}, nil
}

func parseListProductsToolInput(input map[string]any) (listProductsToolInput, error) {
	page, err := listProductsToolNonNegativeInt(firstInventoryToolValue(input, "page"))
	if err != nil {
		return listProductsToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	pageSize, err := listProductsToolNonNegativeInt(firstInventoryToolValue(input, "page_size", "limit"))
	if err != nil {
		return listProductsToolInput{}, fmt.Errorf("invalid page_size: %w", err)
	}
	if pageSize == 0 {
		pageSize = defaultListProductsToolLimit
	}
	if pageSize > maxListProductsToolPageSize {
		pageSize = maxListProductsToolPageSize
	}

	lowStockOnly, err := inventoryToolBool(firstInventoryToolValue(input, "low_stock_only", "low_stock"), false)
	if err != nil {
		return listProductsToolInput{}, fmt.Errorf("invalid low_stock_only: %w", err)
	}

	return listProductsToolInput{
		Search:       strings.TrimSpace(stockToolString(firstInventoryToolValue(input, "search", "query", "q"))),
		LowStockOnly: lowStockOnly,
		Page:         page,
		PageSize:     pageSize,
	}, nil
}

func filterInventoryReportItems(items []domain.InventoryReportItem, filter listProductsToolInput) []domain.InventoryReportItem {
	search := strings.ToLower(filter.Search)
	out := make([]domain.InventoryReportItem, 0, len(items))
	for _, item := range items {
		if filter.LowStockOnly && !item.IsLowStock {
			continue
		}
		if search != "" {
			name := strings.ToLower(item.ProductName)
			sku := strings.ToLower(item.SKU)
			if !strings.Contains(name, search) && !strings.Contains(sku, search) {
				continue
			}
		}
		out = append(out, item)
	}
	return out
}

func paginateInventoryReportItems(items []domain.InventoryReportItem, page, pageSize int) []domain.InventoryReportItem {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultListProductsToolLimit
	}
	start := (page - 1) * pageSize
	if start >= len(items) {
		return nil
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func inventoryReportToolPayload(item domain.InventoryReportItem) map[string]any {
	return map[string]any{
		"id":                  item.ProductID,
		"product_id":          item.ProductID,
		"name":                item.ProductName,
		"sku":                 item.SKU,
		"unit":                item.Unit,
		"stock_quantity":      item.StockQuantity,
		"low_stock_threshold": item.LowStockThreshold,
		"reorder_quantity":    item.ReorderQuantity,
		"cost_price":          item.CostPrice,
		"sale_price":          item.SalePrice,
		"stock_value":         item.StockValue,
		"is_low_stock":        item.IsLowStock,
	}
}

func firstInventoryToolValue(input map[string]any, keys ...string) any {
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

func listProductsToolNonNegativeInt(value any) (int, error) {
	out, err := stockToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	if out > math.MaxInt {
		return 0, errors.New("out of range")
	}
	return int(out), nil
}

func inventoryToolBool(value any, defaultValue bool) (bool, error) {
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
		case "true", "t", "yes", "y", "1", "low", "low_stock":
			return true, nil
		case "false", "f", "no", "n", "0", "all":
			return false, nil
		default:
			return false, errors.New("must be true or false")
		}
	default:
		return false, fmt.Errorf("unsupported value %T", value)
	}
}

func stockToolInt64(value any) (int64, error) {
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

func stockToolFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case json.Number:
		return v.Float64()
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return 0, nil
		}
		return strconv.ParseFloat(raw, 64)
	default:
		return 0, fmt.Errorf("unsupported value %T", value)
	}
}

func stockToolString(value any) string {
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

func formatStockQuantity(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
