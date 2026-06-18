package purchase_orders

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/purchase_orders/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	createPurchaseOrderToolName  = "erp.create_purchase_order"
	approvePurchaseOrderToolName = "erp.approve_purchase_order"
)

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case createPurchaseOrderToolName:
		output, err := m.executeCreatePurchaseOrderTool(ctx, input)
		return output, true, err
	case approvePurchaseOrderToolName:
		output, err := m.executeApprovePurchaseOrderTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeCreatePurchaseOrderTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseCreatePurchaseOrderToolInput(input)
	if err != nil {
		return nil, err
	}

	ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
	id, err := m.service.CreatePurchaseOrder(ctx, ownerID, parsed, authUser.UserID)
	if err != nil {
		return nil, err
	}

	subtotal := purchaseOrderToolSubtotal(parsed.Items)
	return map[string]any{
		"purchase_order": map[string]any{
			"id":                id,
			"supplier_id":       parsed.SupplierID,
			"po_number":         parsed.PONumber,
			"status":            string(domain.POStatusDraft),
			"subtotal":          subtotal,
			"tax_amount":        parsed.TaxAmount,
			"total":             subtotal + parsed.TaxAmount,
			"expected_delivery": parsed.ExpectedDelivery,
			"notes":             parsed.Notes,
			"items":             purchaseOrderToolItemsOutput(parsed.Items),
			"item_count":        len(parsed.Items),
		},
	}, nil
}

func (m *Module) executeApprovePurchaseOrderTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	poID, err := parseApprovePurchaseOrderToolInput(input)
	if err != nil {
		return nil, err
	}

	ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
	if err := m.service.ApprovePurchaseOrder(ctx, ownerID, poID); err != nil {
		return nil, err
	}

	return map[string]any{
		"purchase_order": map[string]any{
			"id":       poID,
			"status":   string(domain.POStatusApproved),
			"approved": true,
		},
	}, nil
}

func parseCreatePurchaseOrderToolInput(input map[string]any) (domain.PurchaseOrderInput, error) {
	supplierID, err := purchaseOrderToolPositiveInt64(firstPurchaseOrderToolValue(input, "supplier_id"))
	if err != nil {
		return domain.PurchaseOrderInput{}, fmt.Errorf("invalid supplier_id: %w", err)
	}

	poNumber := strings.TrimSpace(purchaseOrderToolString(firstPurchaseOrderToolValue(input, "po_number", "po", "purchase_order_number", "number")))
	if poNumber == "" {
		return domain.PurchaseOrderInput{}, errors.New("po_number is required")
	}

	taxAmount, err := purchaseOrderToolNonNegativeFloat64(firstPurchaseOrderToolValue(input, "tax_amount", "tax"))
	if err != nil {
		return domain.PurchaseOrderInput{}, fmt.Errorf("invalid tax_amount: %w", err)
	}

	expectedDelivery, err := purchaseOrderToolDateString(firstPurchaseOrderToolValue(input, "expected_delivery", "expected_delivery_date", "delivery_date"))
	if err != nil {
		return domain.PurchaseOrderInput{}, fmt.Errorf("invalid expected_delivery: %w", err)
	}

	items, err := purchaseOrderToolItems(firstPurchaseOrderToolValue(input, "items", "line_items"))
	if err != nil {
		return domain.PurchaseOrderInput{}, fmt.Errorf("invalid items: %w", err)
	}

	parsed := domain.PurchaseOrderInput{
		SupplierID:       supplierID,
		PONumber:         poNumber,
		TaxAmount:        taxAmount,
		ExpectedDelivery: expectedDelivery,
		Notes:            strings.TrimSpace(purchaseOrderToolString(firstPurchaseOrderToolValue(input, "notes"))),
		Items:            items,
	}
	if err := parsed.Validate(); err != nil {
		return domain.PurchaseOrderInput{}, err
	}
	return parsed, nil
}

func parseApprovePurchaseOrderToolInput(input map[string]any) (int64, error) {
	poID, err := purchaseOrderToolPositiveInt64(firstPurchaseOrderToolValue(input, "purchase_order_id", "po_id", "id"))
	if err != nil {
		return 0, fmt.Errorf("invalid purchase_order_id: %w", err)
	}
	return poID, nil
}

func firstPurchaseOrderToolValue(input map[string]any, keys ...string) any {
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

func purchaseOrderToolItems(value any) ([]domain.PurchaseOrderItemInput, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case []domain.PurchaseOrderItemInput:
		return v, nil
	case []map[string]any:
		items := make([]domain.PurchaseOrderItemInput, 0, len(v))
		for idx, item := range v {
			parsed, err := purchaseOrderToolItem(item)
			if err != nil {
				return nil, fmt.Errorf("item %d: %w", idx+1, err)
			}
			items = append(items, parsed)
		}
		return items, nil
	case []any:
		items := make([]domain.PurchaseOrderItemInput, 0, len(v))
		for idx, item := range v {
			itemMap, ok := item.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("item %d must be an object", idx+1)
			}
			parsed, err := purchaseOrderToolItem(itemMap)
			if err != nil {
				return nil, fmt.Errorf("item %d: %w", idx+1, err)
			}
			items = append(items, parsed)
		}
		return items, nil
	default:
		return nil, fmt.Errorf("must be an array, got %T", value)
	}
}

func purchaseOrderToolItem(input map[string]any) (domain.PurchaseOrderItemInput, error) {
	productID, err := purchaseOrderToolPositiveInt64(firstPurchaseOrderToolValue(input, "product_id"))
	if err != nil {
		return domain.PurchaseOrderItemInput{}, fmt.Errorf("invalid product_id: %w", err)
	}

	quantityOrdered, err := purchaseOrderToolPositiveFloat64(firstPurchaseOrderToolValue(input, "quantity_ordered", "quantity", "qty", "units"))
	if err != nil {
		return domain.PurchaseOrderItemInput{}, fmt.Errorf("invalid quantity_ordered: %w", err)
	}

	unitCost, err := purchaseOrderToolNonNegativeFloat64(firstPurchaseOrderToolValue(input, "unit_cost", "cost", "price"))
	if err != nil {
		return domain.PurchaseOrderItemInput{}, fmt.Errorf("invalid unit_cost: %w", err)
	}

	parsed := domain.PurchaseOrderItemInput{
		ProductID:       productID,
		QuantityOrdered: quantityOrdered,
		UnitCost:        unitCost,
	}
	if err := parsed.Validate(); err != nil {
		return domain.PurchaseOrderItemInput{}, err
	}
	return parsed, nil
}

func purchaseOrderToolPositiveInt64(value any) (int64, error) {
	out, err := purchaseOrderToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out <= 0 {
		return 0, errors.New("must be positive")
	}
	return out, nil
}

func purchaseOrderToolInt64(value any) (int64, error) {
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
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, errors.New("must be finite")
		}
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

func purchaseOrderToolPositiveFloat64(value any) (float64, error) {
	out, err := purchaseOrderToolFloat64(value)
	if err != nil {
		return 0, err
	}
	if out <= 0 {
		return 0, errors.New("must be positive")
	}
	return out, nil
}

func purchaseOrderToolNonNegativeFloat64(value any) (float64, error) {
	out, err := purchaseOrderToolFloat64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func purchaseOrderToolFloat64(value any) (float64, error) {
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

func purchaseOrderToolDateString(value any) (string, error) {
	switch v := value.(type) {
	case nil:
		return "", nil
	case time.Time:
		if v.IsZero() {
			return "", nil
		}
		return v.Format("2006-01-02"), nil
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return "", nil
		}
		for _, layout := range []string{"2006-01-02", time.RFC3339, time.RFC3339Nano} {
			parsed, err := time.Parse(layout, raw)
			if err == nil {
				return parsed.Format("2006-01-02"), nil
			}
		}
		return "", errors.New("must be YYYY-MM-DD or RFC3339")
	default:
		return "", fmt.Errorf("unsupported value %T", value)
	}
}

func purchaseOrderToolString(value any) string {
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

func purchaseOrderToolSubtotal(items []domain.PurchaseOrderItemInput) float64 {
	var subtotal float64
	for _, item := range items {
		subtotal += item.QuantityOrdered * item.UnitCost
	}
	return subtotal
}

func purchaseOrderToolItemsOutput(items []domain.PurchaseOrderItemInput) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]any{
			"product_id":       item.ProductID,
			"quantity_ordered": item.QuantityOrdered,
			"unit_cost":        item.UnitCost,
			"total_cost":       item.QuantityOrdered * item.UnitCost,
		})
	}
	return out
}
