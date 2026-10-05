package orders

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos/orders/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

func TestParseListOrdersToolInput(t *testing.T) {
	parsed, err := parseListOrdersToolInput(map[string]any{
		"session":   "3201",
		"status":    "ready",
		"date_from": "2026-06-18T10:30:00Z",
		"date_to":   "2026-06-19",
		"page":      "2",
		"page_size": 100,
	})
	if err != nil {
		t.Fatalf("parseListOrdersToolInput returned error: %v", err)
	}
	if parsed.SessionID != 3201 {
		t.Fatalf("SessionID = %d, want 3201", parsed.SessionID)
	}
	if parsed.Status != string(domain.OrderStatusReady) {
		t.Fatalf("Status = %q, want ready", parsed.Status)
	}
	if parsed.DateFrom != "2026-06-18" {
		t.Fatalf("DateFrom = %q, want 2026-06-18", parsed.DateFrom)
	}
	if parsed.DateTo != "2026-06-19" {
		t.Fatalf("DateTo = %q, want 2026-06-19", parsed.DateTo)
	}
	if parsed.Page != 2 {
		t.Fatalf("Page = %d, want 2", parsed.Page)
	}
	if parsed.Limit != maxListOrdersToolLimit {
		t.Fatalf("Limit = %d, want cap %d", parsed.Limit, maxListOrdersToolLimit)
	}
}

func TestParseListOrdersToolInputUsesDefaults(t *testing.T) {
	parsed, err := parseListOrdersToolInput(map[string]any{})
	if err != nil {
		t.Fatalf("parseListOrdersToolInput returned error: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d, want 1", parsed.Page)
	}
	if parsed.Limit != defaultListOrdersLimit {
		t.Fatalf("Limit = %d, want %d", parsed.Limit, defaultListOrdersLimit)
	}
}

func TestParseListOrdersToolInputRejectsInvalidStatus(t *testing.T) {
	_, err := parseListOrdersToolInput(map[string]any{"status": "stuck"})
	if err == nil || !strings.Contains(err.Error(), "status") {
		t.Fatalf("error = %v, want status validation", err)
	}
}

func TestParseListOrdersToolInputRejectsInvalidDateRange(t *testing.T) {
	_, err := parseListOrdersToolInput(map[string]any{
		"date_from": "2026-06-19",
		"date_to":   "2026-06-18",
	})
	if err == nil || !strings.Contains(err.Error(), "date_to") {
		t.Fatalf("error = %v, want date range validation", err)
	}
}

func TestParseGetOrderToolInputAcceptsAliases(t *testing.T) {
	for _, input := range []map[string]any{
		{"order_id": "17"},
		{"order": "#17"},
		{"id": "order-17"},
		{"order_id": 17},
	} {
		got, err := parseGetOrderToolInput(input)
		if err != nil {
			t.Fatalf("parseGetOrderToolInput(%v) returned error: %v", input, err)
		}
		if got != 17 {
			t.Fatalf("parseGetOrderToolInput(%v) = %d, want 17", input, got)
		}
	}
}

func TestParseGetOrderToolInputRejectsMissingOrder(t *testing.T) {
	_, err := parseGetOrderToolInput(map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "order_id is required") {
		t.Fatalf("error = %v, want missing order_id validation", err)
	}
}

func TestOrderToolPayload(t *testing.T) {
	paidAt := time.Date(2026, 6, 19, 12, 30, 0, 0, time.UTC)
	payload := orderToolPayload(domain.Order{
		ID:             17,
		SessionID:      3201,
		TableID:        4,
		LocationID:     8,
		OrderNumber:    "POS-17",
		Type:           domain.OrderTypeDineIn,
		Status:         domain.OrderStatusPaid,
		CustomerID:     99,
		CustomerName:   "Maya Haddad",
		StaffID:        5,
		Subtotal:       100,
		DiscountAmount: 10,
		TaxAmount:      9.9,
		Total:          99.9,
		PaymentMethod:  domain.PaymentMethodCard,
		PaymentStatus:  domain.PaymentStatusPaid,
		AmountPaid:     99.9,
		ChangeDue:      0,
		CreatedAt:      time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC),
		PaidAt:         &paidAt,
		Items:          []domain.OrderItem{{ID: 1}, {ID: 2}},
	})
	if payload["order_number"] != "POS-17" {
		t.Fatalf("order_number = %v, want POS-17", payload["order_number"])
	}
	if payload["item_count"] != 2 {
		t.Fatalf("item_count = %v, want 2", payload["item_count"])
	}
	if payload["created_at"] != "2026-06-19T12:00:00Z" {
		t.Fatalf("created_at = %v, want formatted timestamp", payload["created_at"])
	}
	if payload["paid_at"] != "2026-06-19T12:30:00Z" {
		t.Fatalf("paid_at = %v, want formatted timestamp", payload["paid_at"])
	}
}

func TestOrderDetailAndItemToolPayload(t *testing.T) {
	voidedAt := time.Date(2026, 6, 19, 13, 0, 0, 0, time.UTC)
	order := domain.Order{
		ID:              17,
		OrderNumber:     "POS-17",
		Status:          domain.OrderStatusVoided,
		PaymentStatus:   domain.PaymentStatusPaid,
		Notes:           "Customer changed order",
		KitchenNotes:    "No onions",
		TerminalID:      3,
		BranchID:        8,
		TipAmount:       4.25,
		VoidReason:      "Mistake",
		VoidedBy:        9,
		VoidedAt:        &voidedAt,
		ExchangeOrderID: 18,
		CreatedAt:       time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2026, 6, 19, 12, 45, 0, 0, time.UTC),
		Items: []domain.OrderItem{{
			ID:             5,
			OrderID:        17,
			MenuItemID:     44,
			Name:           "Burger",
			SKU:            "BRG",
			Quantity:       2,
			UnitPrice:      8.5,
			DiscountPct:    10,
			DiscountAmount: 1.7,
			TaxRate:        11,
			Total:          16.98,
			Notes:          "Medium",
			Status:         domain.ItemStatusServed,
			CreatedAt:      time.Date(2026, 6, 19, 12, 1, 0, 0, time.UTC),
		}},
	}

	detail := orderDetailToolPayload(order)
	if detail["notes"] != "Customer changed order" || detail["kitchen_notes"] != "No onions" {
		t.Fatalf("detail notes missing: %#v", detail)
	}
	if detail["voided_at"] != "2026-06-19T13:00:00Z" {
		t.Fatalf("voided_at = %v, want formatted timestamp", detail["voided_at"])
	}
	if detail["exchange_order_id"] != int64(18) {
		t.Fatalf("exchange_order_id = %v, want 18", detail["exchange_order_id"])
	}

	item := orderItemToolPayload(order.Items[0])
	if item["name"] != "Burger" || item["status"] != "served" {
		t.Fatalf("item payload missing display fields: %#v", item)
	}
	if item["created_at"] != "2026-06-19T12:01:00Z" {
		t.Fatalf("item created_at = %v, want formatted timestamp", item["created_at"])
	}
}

func TestParseRefundOrderToolInput(t *testing.T) {
	parsed, err := parseRefundOrderToolInput(map[string]any{
		"order_id": "17",
		"amount":   "12.50",
		"reason":   "Customer returned item",
	})
	if err != nil {
		t.Fatalf("parseRefundOrderToolInput returned error: %v", err)
	}
	if parsed.OrderID != 17 {
		t.Fatalf("OrderID = %d, want 17", parsed.OrderID)
	}
	if parsed.Amount != 12.50 {
		t.Fatalf("Amount = %v, want 12.50", parsed.Amount)
	}
	if parsed.Reason != "Customer returned item" {
		t.Fatalf("Reason = %q, want Customer returned item", parsed.Reason)
	}
}

func TestParseRefundOrderToolInputRequiresReason(t *testing.T) {
	_, err := parseRefundOrderToolInput(map[string]any{
		"order_id": 17,
		"amount":   12.50,
	})
	if err == nil || !strings.Contains(err.Error(), "reason is required") {
		t.Fatalf("error = %v, want reason validation", err)
	}
}

func TestValidateRefundableOrderRejectsOverRefund(t *testing.T) {
	err := validateRefundableOrder(&domain.Order{
		ID:            17,
		Status:        domain.OrderStatusPaid,
		PaymentStatus: domain.PaymentStatusPaid,
		AmountPaid:    10,
		Total:         10,
	}, 12.50)
	if err == nil || !strings.Contains(err.Error(), "cannot exceed") {
		t.Fatalf("error = %v, want over-refund validation", err)
	}
}

func TestValidateRefundableOrderRejectsUnpaidOrder(t *testing.T) {
	err := validateRefundableOrder(&domain.Order{
		ID:            17,
		Status:        domain.OrderStatusServed,
		PaymentStatus: domain.PaymentStatusUnpaid,
		AmountPaid:    0,
		Total:         10,
	}, 5)
	if err == nil || !strings.Contains(err.Error(), "paid amount") {
		t.Fatalf("error = %v, want paid amount validation", err)
	}
}

func TestValidateRefundableOrderRejectsAlreadyRefunded(t *testing.T) {
	err := validateRefundableOrder(&domain.Order{
		ID:            17,
		Status:        domain.OrderStatusRefunded,
		PaymentStatus: domain.PaymentStatusRefunded,
		AmountPaid:    10,
		Total:         10,
	}, 5)
	if err == nil || !strings.Contains(err.Error(), "already refunded") {
		t.Fatalf("error = %v, want already-refunded validation", err)
	}
}

func toolUserWithoutPOSGrants() *authPkg.AuthUser {
	return &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppPOS,
				Licensed:    true,
				HasAccess:   true,
				Permissions: []string{"pos:menu:view"},
			}},
		},
	}
}

func TestCaretToolsDenyWithoutPOSPermission(t *testing.T) {
	m := &Module{}
	ctx := authPkg.WithAuthUser(context.Background(), toolUserWithoutPOSGrants())

	if _, err := m.executeListOrdersTool(ctx, map[string]any{}); err == nil ||
		!strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("list orders without grant returned %v, want permission denied", err)
	}
	if _, err := m.executeGetOrderTool(ctx, map[string]any{"order_id": 1}); err == nil ||
		!strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("get order without grant returned %v, want permission denied", err)
	}
	if _, err := m.executeRefundOrderTool(ctx, map[string]any{"order_id": 1, "amount": 5}); err == nil ||
		!strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("refund without grant returned %v, want permission denied", err)
	}
}

func TestCaretToolsDenyUnauthenticated(t *testing.T) {
	m := &Module{}
	if _, err := m.executeListOrdersTool(context.Background(), map[string]any{}); err == nil {
		t.Fatal("expected authentication error without user")
	}
}
