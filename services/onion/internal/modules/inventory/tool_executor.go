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

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	if name != "erp.set_stock" {
		return nil, false, nil
	}
	output, err := m.executeSetStockTool(ctx, input)
	return output, true, err
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

	ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
	report, err := m.service.GetInventoryReport(ctx, ownerID)
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
		if err := m.service.AdjustInventory(ctx, ownerID, domain.InventoryAdjustInput{
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
