package pos_orders

import (
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
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
