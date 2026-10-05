package onlineorders

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/erp/onlineorders/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/pagination"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listStoreOrdersToolName         = "store.list_orders"
	getStoreOrderToolName           = "store.get_order"
	defaultListStoreOrdersToolLimit = int64(20)
	maxListStoreOrdersToolLimit     = int64(50)
	storeOrderToolTimestampDate     = "2006-01-02T15:04:05Z"
)

type listStoreOrdersToolInput struct {
	Query    string
	Status   string
	MinTotal float64
	Page     int64
	Limit    int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listStoreOrdersToolName:
		output, err := m.executeListStoreOrdersTool(ctx, input)
		return output, true, err
	case getStoreOrderToolName:
		output, err := m.executeGetStoreOrderTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListStoreOrdersTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, errors.New("permission denied")
	}

	parsed, err := parseListStoreOrdersToolInput(input)
	if err != nil {
		return nil, err
	}

	orders, err := m.service.List(ctx, businessID)
	if err != nil {
		return nil, err
	}
	orders = filterStoreOrdersForTool(orders, parsed)

	start := pagination.SliceStart(parsed.Page, parsed.Limit, len(orders))
	if start > len(orders) {
		start = len(orders)
	}
	end := start + int(parsed.Limit)
	if end > len(orders) {
		end = len(orders)
	}

	payloads := make([]map[string]any, 0, end-start)
	for _, order := range orders[start:end] {
		payloads = append(payloads, storeOrderToolPayload(order))
	}

	return map[string]any{
		"orders":    payloads,
		"total":     len(orders),
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.Limit,
		"has_more":  end < len(orders),
		"filters": map[string]any{
			"query":     parsed.Query,
			"status":    parsed.Status,
			"min_total": parsed.MinTotal,
		},
	}, nil
}

func (m *Module) executeGetStoreOrderTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, errors.New("permission denied")
	}

	orderID, err := storeOrderToolInt64(firstStoreOrderToolValue(input, "order_id", "id"))
	if err != nil {
		return nil, fmt.Errorf("invalid order_id: %w", err)
	}
	if orderID <= 0 {
		return nil, errors.New("order_id is required")
	}

	detail, err := m.service.GetDetail(ctx, businessID, orderID)
	if err != nil {
		return nil, err
	}

	items := make([]map[string]any, 0, len(detail.Items))
	for _, item := range detail.Items {
		items = append(items, storeOrderItemToolPayload(item))
	}

	return map[string]any{
		"order": storeOrderToolPayload(detail.Order),
		"items": items,
	}, nil
}

func parseListStoreOrdersToolInput(input map[string]any) (listStoreOrdersToolInput, error) {
	query := strings.TrimSpace(storeOrderToolString(firstStoreOrderToolValue(input, "query", "q", "search", "customer")))
	status := strings.TrimSpace(storeOrderToolString(firstStoreOrderToolValue(input, "status")))

	minTotal, err := storeOrderToolNonNegativeFloat64(firstStoreOrderToolValue(input, "min_total", "minimum_total"))
	if err != nil {
		return listStoreOrdersToolInput{}, fmt.Errorf("invalid min_total: %w", err)
	}

	page, err := storeOrderToolNonNegativeInt64(firstStoreOrderToolValue(input, "page"))
	if err != nil {
		return listStoreOrdersToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	limit, err := storeOrderToolNonNegativeInt64(firstStoreOrderToolValue(input, "limit", "page_size"))
	if err != nil {
		return listStoreOrdersToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListStoreOrdersToolLimit
	}
	if limit > maxListStoreOrdersToolLimit {
		limit = maxListStoreOrdersToolLimit
	}

	return listStoreOrdersToolInput{
		Query:    query,
		Status:   status,
		MinTotal: minTotal,
		Page:     page,
		Limit:    limit,
	}, nil
}

func filterStoreOrdersForTool(orders []domain.StoreOrder, input listStoreOrdersToolInput) []domain.StoreOrder {
	if input.Query == "" && input.Status == "" && input.MinTotal == 0 {
		return orders
	}

	query := strings.ToLower(input.Query)
	out := make([]domain.StoreOrder, 0, len(orders))
	for _, order := range orders {
		if input.Status != "" && !strings.EqualFold(order.Status, input.Status) {
			continue
		}
		if input.MinTotal > 0 && order.Total < input.MinTotal {
			continue
		}
		if query != "" && !storeOrderToolMatchesQuery(order, query) {
			continue
		}
		out = append(out, order)
	}
	return out
}

func storeOrderToolMatchesQuery(order domain.StoreOrder, query string) bool {
	return strings.Contains(strings.ToLower(order.CustomerEmail), query) ||
		strings.Contains(strings.ToLower(order.CustomerName), query) ||
		strings.Contains(strings.ToLower(order.CustomerPhone), query) ||
		strings.Contains(strings.ToLower(order.ShippingAddress), query) ||
		strings.Contains(strings.ToLower(order.Notes), query)
}

func storeOrderToolPayload(order domain.StoreOrder) map[string]any {
	return map[string]any{
		"id":               order.ID,
		"owner_id":         order.OwnerID,
		"business_id":      order.BusinessID,
		"customer_email":   order.CustomerEmail,
		"customer_name":    order.CustomerName,
		"customer_phone":   order.CustomerPhone,
		"status":           order.Status,
		"subtotal":         order.Subtotal,
		"tax":              order.Tax,
		"shipping_cost":    order.ShippingCost,
		"total":            order.Total,
		"shipping_address": order.ShippingAddress,
		"notes":            order.Notes,
		"created_at":       formatStoreOrderToolTime(order.CreatedAt),
		"updated_at":       formatStoreOrderToolTime(order.UpdatedAt),
	}
}

func storeOrderItemToolPayload(item domain.StoreOrderItem) map[string]any {
	return map[string]any{
		"id":         item.ID,
		"order_id":   item.OrderID,
		"product_id": item.ProductID,
		"name":       item.Name,
		"quantity":   item.Quantity,
		"price":      item.Price,
		"total":      item.Total,
		"created_at": formatStoreOrderToolTime(item.CreatedAt),
	}
}

func firstStoreOrderToolValue(input map[string]any, keys ...string) any {
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

func storeOrderToolNonNegativeInt64(value any) (int64, error) {
	out, err := storeOrderToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func storeOrderToolInt64(value any) (int64, error) {
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

func storeOrderToolNonNegativeFloat64(value any) (float64, error) {
	out, err := storeOrderToolFloat64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func storeOrderToolFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case float64:
		return v, nil
	case json.Number:
		return strconv.ParseFloat(v.String(), 64)
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

func storeOrderToolString(value any) string {
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

func formatStoreOrderToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(storeOrderToolTimestampDate)
}
