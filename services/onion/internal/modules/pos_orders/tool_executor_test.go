package pos_orders

import (
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
)

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
