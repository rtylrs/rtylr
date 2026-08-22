package domain

import (
	"errors"
	"testing"
)

func TestErrorsAreDistinct(t *testing.T) {
	all := []error{
		ErrSessionAlreadyOpen,
		ErrSessionNotFound,
		ErrSessionNotOpen,
		ErrOrderNotFound,
		ErrOrderAlreadyPaid,
		ErrPaymentInsufficient,
		ErrLocationNotFound,
		ErrTableNotFound,
	}
	for i, a := range all {
		for j, b := range all {
			if i != j && errors.Is(a, b) {
				t.Errorf("%v should not be Is %v", a, b)
			}
		}
	}
}

func TestTableStatus_KnownValues(t *testing.T) {
	for _, s := range []TableStatus{
		TableStatusAvailable, TableStatusOccupied, TableStatusReserved, TableStatusCleaning,
	} {
		if string(s) == "" {
			t.Errorf("table status should be non-empty")
		}
	}
	if TableStatusAvailable == TableStatusOccupied {
		t.Error("constants must be distinct")
	}
}

func TestSessionStatus_KnownValues(t *testing.T) {
	if SessionStatusOpen == SessionStatusClosed {
		t.Error("open and closed must be distinct")
	}
	if string(SessionStatusOpen) != "open" || string(SessionStatusClosed) != "closed" {
		t.Errorf("status string mismatch: open=%q closed=%q", SessionStatusOpen, SessionStatusClosed)
	}
}

func TestOrderType_KnownValues(t *testing.T) {
	all := map[OrderType]string{
		OrderTypeDineIn:   "dine_in",
		OrderTypeTakeaway: "takeaway",
		OrderTypeDelivery: "delivery",
		OrderTypeOnline:   "online",
	}
	for typ, want := range all {
		if string(typ) != want {
			t.Errorf("OrderType %v != %q", typ, want)
		}
		if !IsOrderType(typ) {
			t.Errorf("OrderType %v should be valid", typ)
		}
	}
	if IsOrderType(OrderType("pickup")) {
		t.Error("unknown order type should be invalid")
	}
}

func TestOrderStatus_KnownValues(t *testing.T) {
	expected := map[OrderStatus]string{
		OrderStatusPending:   "pending",
		OrderStatusConfirmed: "confirmed",
		OrderStatusPreparing: "preparing",
		OrderStatusReady:     "ready",
		OrderStatusServed:    "served",
		OrderStatusPaid:      "paid",
		OrderStatusCancelled: "cancelled",
		OrderStatusRefunded:  "refunded",
		OrderStatusVoided:    "voided",
		OrderStatusHeld:      "held",
	}
	for s, want := range expected {
		if string(s) != want {
			t.Errorf("OrderStatus %v != %q", s, want)
		}
	}
}

func TestPaymentMethodAndStatusAreDistinct(t *testing.T) {
	if PaymentMethodCash == PaymentMethodCard || PaymentMethodCash == PaymentMethodOnline {
		t.Error("payment methods should be distinct")
	}
	if PaymentStatusPaid == PaymentStatusUnpaid || PaymentStatusPaid == PaymentStatusRefunded {
		t.Error("payment statuses should be distinct")
	}
}
