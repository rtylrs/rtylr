package purchase_orders

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/erp/purchase_orders/internal/domain"
)

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

func normalizePurchaseOrderToolStatus(value any) string {
	status := strings.ToLower(strings.TrimSpace(purchaseOrderToolString(value)))
	if status == "" {
		return ""
	}
	status = strings.ReplaceAll(status, "-", "_")
	status = strings.ReplaceAll(status, " ", "_")
	switch domain.PurchaseOrderStatus(status) {
	case domain.POStatusDraft,
		domain.POStatusSent,
		domain.POStatusApproved,
		domain.POStatusPartial,
		domain.POStatusReceived,
		domain.POStatusCancelled:
		return status
	default:
		return ""
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

func purchaseOrderToolPayload(po domain.PurchaseOrder) map[string]any {
	return map[string]any{
		"id":                po.ID,
		"supplier_id":       po.SupplierID,
		"po_number":         po.PONumber,
		"status":            string(po.Status),
		"subtotal":          po.Subtotal,
		"tax_amount":        po.TaxAmount,
		"total":             po.Total,
		"expected_delivery": formatPurchaseOrderToolDate(po.ExpectedDelivery),
		"received_at":       formatPurchaseOrderToolTime(po.ReceivedAt),
		"notes":             po.Notes,
		"created_by":        po.CreatedBy,
		"created_at":        formatPurchaseOrderToolRequiredTime(po.CreatedAt),
		"updated_at":        formatPurchaseOrderToolRequiredTime(po.UpdatedAt),
		"item_count":        len(po.Items),
		"items":             purchaseOrderDomainItemsOutput(po.Items),
	}
}

func purchaseOrderDomainItemsOutput(items []domain.PurchaseOrderItem) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]any{
			"id":                item.ID,
			"purchase_order_id": item.PurchaseOrderID,
			"product_id":        item.ProductID,
			"quantity_ordered":  item.QuantityOrdered,
			"quantity_received": item.QuantityReceived,
			"unit_cost":         item.UnitCost,
			"total_cost":        item.TotalCost,
		})
	}
	return out
}

func formatPurchaseOrderToolDate(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02")
}

func formatPurchaseOrderToolTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02T15:04:05Z")
}

func formatPurchaseOrderToolRequiredTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02T15:04:05Z")
}
