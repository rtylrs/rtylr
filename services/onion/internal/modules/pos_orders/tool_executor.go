package pos_orders

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/pagination"
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
