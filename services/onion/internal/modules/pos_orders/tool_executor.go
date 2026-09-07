package pos_orders

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listOrdersToolName     = "pos.list_orders"
	getOrderToolName       = "pos.get_order"
	refundOrderToolName    = "pos.refund_order"
	refundAmountTolerance  = 0.000001
	defaultListOrdersLimit = int64(20)
	maxListOrdersToolLimit = int64(50)
	posToolCalendarDate    = "2006-01-02"
	posToolTimestampDate   = "2006-01-02T15:04:05Z"
)

type listOrdersToolInput struct {
	SessionID int64
	Status    string
	DateFrom  string
	DateTo    string
	Page      int64
	Limit     int64
}

type refundOrderToolInput struct {
	OrderID int64
	Amount  float64
	Reason  string
}

// requirePOSToolPermission enforces the same POS grant check as the HTTP
// handlers so the Caret tool surface cannot bypass route authorization.
func requirePOSToolPermission(authUser *authPkg.AuthUser, permission string) error {
	if !authPkg.HasAppPermission(authUser, authorization.AppPOS, permission) ||
		!authPkg.HasAnyAppScope(authUser, authorization.AppPOS) {
		return errors.New("permission denied")
	}
	return nil
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listOrdersToolName:
		output, err := m.executeListOrdersTool(ctx, input)
		return output, true, err
	case getOrderToolName:
		output, err := m.executeGetOrderTool(ctx, input)
		return output, true, err
	case refundOrderToolName:
		output, err := m.executeRefundOrderTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListOrdersTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}
	if err := requirePOSToolPermission(authUser, "pos:orders:view"); err != nil {
		return nil, err
	}

	parsed, err := parseListOrdersToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, errors.New("cannot determine business context")
	}
	orders, err := m.service.ListOrders(ctx, businessID, parsed.SessionID, parsed.Status, parsed.DateFrom, parsed.DateTo)
	if err != nil {
		return nil, err
	}

	start := int((parsed.Page - 1) * parsed.Limit)
	if start > len(orders) {
		start = len(orders)
	}
	end := start + int(parsed.Limit)
	if end > len(orders) {
		end = len(orders)
	}

	payloads := make([]map[string]any, 0, end-start)
	for _, order := range orders[start:end] {
		payloads = append(payloads, orderToolPayload(order))
	}

	return map[string]any{
		"orders":    payloads,
		"total":     len(orders),
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.Limit,
		"has_more":  end < len(orders),
		"filters": map[string]any{
			"session_id": parsed.SessionID,
			"status":     parsed.Status,
			"date_from":  parsed.DateFrom,
			"date_to":    parsed.DateTo,
		},
	}, nil
}

func (m *Module) executeGetOrderTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}
	if err := requirePOSToolPermission(authUser, "pos:orders:view"); err != nil {
		return nil, err
	}

	orderID, err := parseGetOrderToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, errors.New("cannot determine business context")
	}
	order, err := m.service.GetOrder(ctx, businessID, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, domain.ErrOrderNotFound
	}

	items := make([]map[string]any, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, orderItemToolPayload(item))
	}

	return map[string]any{
		"order": orderDetailToolPayload(*order),
		"items": items,
	}, nil
}

func (m *Module) executeRefundOrderTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}
	if err := requirePOSToolPermission(authUser, "pos:orders:refund"); err != nil {
		return nil, err
	}

	parsed, err := parseRefundOrderToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, errors.New("cannot determine business context")
	}
	order, err := m.service.GetOrder(ctx, businessID, parsed.OrderID)
	if err != nil {
		return nil, err
	}
	if err := validateRefundableOrder(order, parsed.Amount); err != nil {
		return nil, err
	}

	refundID, err := m.service.ProcessRefund(ctx, businessID, authUser.UserID, domain.RefundInput{
		OrderID: parsed.OrderID,
		Amount:  parsed.Amount,
		Reason:  parsed.Reason,
	})
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"refund": map[string]any{
			"id":                      refundID,
			"order_id":                parsed.OrderID,
			"order_number":            order.OrderNumber,
			"amount":                  parsed.Amount,
			"reason":                  parsed.Reason,
			"previous_status":         string(order.Status),
			"previous_payment_status": string(order.PaymentStatus),
		},
	}, nil
}

func parseListOrdersToolInput(input map[string]any) (listOrdersToolInput, error) {
	sessionID, err := posToolInt64(firstPosToolValue(input, "session_id", "session"))
	if err != nil {
		return listOrdersToolInput{}, fmt.Errorf("invalid session_id: %w", err)
	}
	if sessionID < 0 {
		return listOrdersToolInput{}, errors.New("session_id must be non-negative")
	}

	status := normalizeOrderToolStatus(input["status"])
	if strings.TrimSpace(posToolString(input["status"])) != "" && status == "" {
		return listOrdersToolInput{}, errors.New("status must be one of pending, confirmed, preparing, ready, served, paid, cancelled, refunded, voided, or held")
	}

	dateFrom, err := posToolDate(firstPosToolValue(input, "date_from", "from", "start_date", "created_from"))
	if err != nil {
		return listOrdersToolInput{}, fmt.Errorf("invalid date_from: %w", err)
	}
	dateTo, err := posToolDate(firstPosToolValue(input, "date_to", "to", "end_date", "created_to"))
	if err != nil {
		return listOrdersToolInput{}, fmt.Errorf("invalid date_to: %w", err)
	}
	if err := validateOrderToolDateRange(dateFrom, dateTo); err != nil {
		return listOrdersToolInput{}, err
	}

	page, err := posToolInt64(firstPosToolValue(input, "page"))
	if err != nil {
		return listOrdersToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page < 0 {
		return listOrdersToolInput{}, errors.New("page must be non-negative")
	}
	if page == 0 {
		page = 1
	}

	limit, err := posToolInt64(firstPosToolValue(input, "limit", "page_size"))
	if err != nil {
		return listOrdersToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit < 0 {
		return listOrdersToolInput{}, errors.New("limit must be non-negative")
	}
	if limit == 0 {
		limit = defaultListOrdersLimit
	}
	if limit > maxListOrdersToolLimit {
		limit = maxListOrdersToolLimit
	}

	return listOrdersToolInput{
		SessionID: sessionID,
		Status:    status,
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		Page:      page,
		Limit:     limit,
	}, nil
}

func parseGetOrderToolInput(input map[string]any) (int64, error) {
	orderID, err := posOrderIDToolValue(firstPosToolValue(input, "order_id", "order", "id"))
	if err != nil {
		return 0, fmt.Errorf("invalid order_id: %w", err)
	}
	if orderID <= 0 {
		return 0, errors.New("order_id is required")
	}
	return orderID, nil
}

func parseRefundOrderToolInput(input map[string]any) (refundOrderToolInput, error) {
	orderIDValue := input["order_id"]
	if orderIDValue == nil {
		orderIDValue = input["id"]
	}
	orderID, err := posToolInt64(orderIDValue)
	if err != nil {
		return refundOrderToolInput{}, fmt.Errorf("invalid order_id: %w", err)
	}
	if orderID <= 0 {
		return refundOrderToolInput{}, errors.New("order_id is required")
	}

	amountValue := input["amount"]
	if amountValue == nil {
		amountValue = input["refund_amount"]
	}
	amount, err := posToolFloat64(amountValue)
	if err != nil {
		return refundOrderToolInput{}, fmt.Errorf("invalid amount: %w", err)
	}
	if amount <= 0 {
		return refundOrderToolInput{}, errors.New("refund amount must be positive")
	}

	reason := strings.TrimSpace(posToolString(input["reason"]))
	if reason == "" {
		reason = strings.TrimSpace(posToolString(input["notes"]))
	}
	if reason == "" {
		return refundOrderToolInput{}, errors.New("reason is required")
	}

	return refundOrderToolInput{
		OrderID: orderID,
		Amount:  amount,
		Reason:  reason,
	}, nil
}

func validateRefundableOrder(order *domain.Order, amount float64) error {
	if order == nil {
		return domain.ErrOrderNotFound
	}
	if order.Status == domain.OrderStatusRefunded || order.PaymentStatus == domain.PaymentStatusRefunded {
		return errors.New("order is already refunded")
	}
	if order.Status == domain.OrderStatusVoided || order.Status == domain.OrderStatusCancelled {
		return errors.New("order cannot be refunded in its current status")
	}
	if order.AmountPaid <= 0 && order.PaymentStatus != domain.PaymentStatusPaid {
		return errors.New("order does not have a paid amount to refund")
	}

	maxRefund := order.AmountPaid
	if maxRefund <= 0 {
		maxRefund = order.Total
	}
	if maxRefund <= 0 {
		return errors.New("order does not have a paid amount to refund")
	}
	if amount-maxRefund > refundAmountTolerance {
		return fmt.Errorf("refund amount cannot exceed %.2f", maxRefund)
	}
	return nil
}

func normalizeOrderToolStatus(value any) string {
	status := strings.ToLower(strings.TrimSpace(posToolString(value)))
	if status == "" {
		return ""
	}
	status = strings.ReplaceAll(status, "-", "_")
	status = strings.ReplaceAll(status, " ", "_")
	switch domain.OrderStatus(status) {
	case domain.OrderStatusPending,
		domain.OrderStatusConfirmed,
		domain.OrderStatusPreparing,
		domain.OrderStatusReady,
		domain.OrderStatusServed,
		domain.OrderStatusPaid,
		domain.OrderStatusCancelled,
		domain.OrderStatusRefunded,
		domain.OrderStatusVoided,
		domain.OrderStatusHeld:
		return status
	default:
		return ""
	}
}

func validateOrderToolDateRange(dateFrom, dateTo string) error {
	if dateFrom == "" || dateTo == "" {
		return nil
	}
	parsedFrom, err := time.Parse(posToolCalendarDate, dateFrom)
	if err != nil {
		return fmt.Errorf("invalid date_from: %w", err)
	}
	parsedTo, err := time.Parse(posToolCalendarDate, dateTo)
	if err != nil {
		return fmt.Errorf("invalid date_to: %w", err)
	}
	if parsedTo.Before(parsedFrom) {
		return errors.New("date_to must be on or after date_from")
	}
	return nil
}

func orderDetailToolPayload(order domain.Order) map[string]any {
	payload := orderToolPayload(order)
	payload["notes"] = order.Notes
	payload["kitchen_notes"] = order.KitchenNotes
	payload["updated_at"] = formatOrderToolTime(order.UpdatedAt)
	payload["terminal_id"] = order.TerminalID
	payload["branch_id"] = order.BranchID
	payload["tip_amount"] = order.TipAmount
	payload["void_reason"] = order.VoidReason
	payload["voided_by"] = order.VoidedBy
	payload["voided_at"] = formatOrderToolOptionalTime(order.VoidedAt)
	payload["exchange_order_id"] = order.ExchangeOrderID
	return payload
}

func orderToolPayload(order domain.Order) map[string]any {
	return map[string]any{
		"id":              order.ID,
		"session_id":      order.SessionID,
		"table_id":        order.TableID,
		"location_id":     order.LocationID,
		"order_number":    order.OrderNumber,
		"type":            string(order.Type),
		"status":          string(order.Status),
		"customer_id":     order.CustomerID,
		"customer_name":   order.CustomerName,
		"staff_id":        order.StaffID,
		"subtotal":        order.Subtotal,
		"discount_amount": order.DiscountAmount,
		"tax_amount":      order.TaxAmount,
		"total":           order.Total,
		"payment_method":  string(order.PaymentMethod),
		"payment_status":  string(order.PaymentStatus),
		"amount_paid":     order.AmountPaid,
		"change_due":      order.ChangeDue,
		"item_count":      len(order.Items),
		"created_at":      formatOrderToolTime(order.CreatedAt),
		"paid_at":         formatOrderToolOptionalTime(order.PaidAt),
	}
}

func orderItemToolPayload(item domain.OrderItem) map[string]any {
	return map[string]any{
		"id":              item.ID,
		"order_id":        item.OrderID,
		"menu_item_id":    item.MenuItemID,
		"name":            item.Name,
		"sku":             item.SKU,
		"quantity":        item.Quantity,
		"unit_price":      item.UnitPrice,
		"discount_pct":    item.DiscountPct,
		"discount_amount": item.DiscountAmount,
		"tax_rate":        item.TaxRate,
		"total":           item.Total,
		"notes":           item.Notes,
		"status":          string(item.Status),
		"created_at":      formatOrderToolTime(item.CreatedAt),
	}
}

func formatOrderToolTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(posToolTimestampDate)
}

func formatOrderToolOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return formatOrderToolTime(*value)
}

func posToolDate(value any) (string, error) {
	switch v := value.(type) {
	case nil:
		return "", nil
	case time.Time:
		if v.IsZero() {
			return "", nil
		}
		return v.Format(posToolCalendarDate), nil
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return "", nil
		}
		for _, layout := range []string{posToolCalendarDate, time.RFC3339, time.RFC3339Nano} {
			parsed, err := time.Parse(layout, raw)
			if err == nil {
				return parsed.Format(posToolCalendarDate), nil
			}
		}
		return "", errors.New("must be YYYY-MM-DD or RFC3339")
	default:
		return "", fmt.Errorf("unsupported value %T", value)
	}
}

func firstPosToolValue(input map[string]any, keys ...string) any {
	for _, key := range keys {
		value, ok := input[key]
		if !ok {
			continue
		}
		if raw, ok := value.(string); ok && strings.TrimSpace(raw) == "" {
			continue
		}
		if number, ok := value.(float64); ok && math.IsNaN(number) {
			continue
		}
		return value
	}
	return nil
}

func posOrderIDToolValue(value any) (int64, error) {
	if raw, ok := value.(string); ok {
		text := strings.TrimSpace(raw)
		if text == "" {
			return 0, nil
		}
		normalized := strings.ToLower(text)
		normalized = strings.TrimPrefix(normalized, "#")
		normalized = strings.TrimPrefix(normalized, "order-")
		if normalized == "" {
			return 0, nil
		}
		return strconv.ParseInt(normalized, 10, 64)
	}
	return posToolInt64(value)
}

func posToolInt64(value any) (int64, error) {
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

func posToolFloat64(value any) (float64, error) {
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

func posToolString(value any) string {
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
