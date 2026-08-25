package application

import (
	"context"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_settings"
)

// ProcessRefund's input guards run before any repository call, so we can
// exercise them with a zero-value Service and verify they reject malformed
// input before touching the database.

func TestProcessRefund_RejectsMissingOrderID(t *testing.T) {
	s := &Service{}
	if _, err := s.ProcessRefund(context.Background(), 1, 1, domain.RefundInput{Amount: 5}); err == nil {
		t.Fatalf("expected error for missing order id")
	} else if !strings.Contains(err.Error(), "order id") {
		t.Fatalf("expected order id error, got: %v", err)
	}
}

func TestProcessRefund_RejectsNonPositiveAmount(t *testing.T) {
	s := &Service{}
	for _, amount := range []float64{0, -1, -0.01} {
		if _, err := s.ProcessRefund(context.Background(), 1, 1, domain.RefundInput{OrderID: 7, Amount: amount}); err == nil {
			t.Fatalf("expected error for amount %v", amount)
		}
	}
}

// ─── CreateOrder input guards ────────────────────────────────────────────────

func TestCreateOrder_RejectsEmptyItems(t *testing.T) {
	s := &Service{}
	if _, err := s.CreateOrder(context.Background(), 1, 1, 1, domain.OrderInput{}); err == nil {
		t.Error("expected error for empty items")
	} else if !strings.Contains(err.Error(), "item") {
		t.Errorf("expected item-related error, got: %v", err)
	}
}

func TestDefaultOrderTypeForProfile(t *testing.T) {
	if got := defaultOrderTypeForProfile(pos_settings.ProfileRetail); got != domain.OrderTypeTakeaway {
		t.Fatalf("retail default = %q, want %q", got, domain.OrderTypeTakeaway)
	}
	if got := defaultOrderTypeForProfile(pos_settings.ProfileQuickService); got != domain.OrderTypeTakeaway {
		t.Fatalf("quick service default = %q, want %q", got, domain.OrderTypeTakeaway)
	}
	if got := defaultOrderTypeForProfile(pos_settings.ProfileRestaurant); got != domain.OrderTypeDineIn {
		t.Fatalf("restaurant default = %q, want %q", got, domain.OrderTypeDineIn)
	}
}

// ─── CreateLocation input guard ──────────────────────────────────────────────

func TestCreateLocation_RejectsEmptyName(t *testing.T) {
	s := &Service{}
	if _, err := s.CreateLocation(context.Background(), 1, domain.LocationInput{}); err == nil {
		t.Error("expected error for empty location name")
	} else if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected name-related error, got: %v", err)
	}
}

// ─── CreateTable input guards ─────────────────────────────────────────────────

func TestCreateTable_RejectsEmptyName(t *testing.T) {
	s := &Service{}
	if _, err := s.CreateTable(context.Background(), 1, domain.TableInput{}); err == nil {
		t.Error("expected error for empty table name")
	}
}

// ─── UpdateTable input guards ─────────────────────────────────────────────────

func TestUpdateTable_RejectsZeroID(t *testing.T) {
	s := &Service{}
	if err := s.UpdateTable(context.Background(), 1, domain.TableInput{Name: "T1"}); err == nil {
		t.Error("expected error for zero table id")
	}
}

func TestUpdateTable_RejectsEmptyName(t *testing.T) {
	s := &Service{}
	if err := s.UpdateTable(context.Background(), 1, domain.TableInput{ID: 1}); err == nil {
		t.Error("expected error for empty name")
	}
}

// ─── UpdateTableStatus / DeleteTable guards ───────────────────────────────────

func TestUpdateTableStatus_RejectsZeroTableID(t *testing.T) {
	s := &Service{}
	if err := s.UpdateTableStatus(context.Background(), 1, 0, domain.TableStatusAvailable); err == nil {
		t.Error("expected error for zero table id")
	}
}

func TestDeleteTable_RejectsZeroTableID(t *testing.T) {
	s := &Service{}
	if err := s.DeleteTable(context.Background(), 1, 0); err == nil {
		t.Error("expected error for zero table id")
	}
}

// ─── OpenSession guard ────────────────────────────────────────────────────────

func TestProcessPayment_RejectsZeroOrderID(t *testing.T) {
	s := &Service{}
	if err := s.ProcessPayment(context.Background(), 1, domain.PaymentInput{OrderID: 0}); err == nil {
		t.Error("expected error for zero order id")
	}
}

// ─── CreatePublicOrder guards ─────────────────────────────────────────────────

func TestCreatePublicOrder_RejectsZeroOwnerID(t *testing.T) {
	s := &Service{}
	if _, _, _, err := s.CreatePublicOrder(context.Background(), 0, PublicOrderInput{
		Items: []PublicOrderItemInput{{Name: "Burger", Quantity: 1}},
	}); err == nil {
		t.Error("expected error for zero owner_id")
	}
}

func TestCreatePublicOrder_RejectsEmptyItems(t *testing.T) {
	s := &Service{}
	if _, _, _, err := s.CreatePublicOrder(context.Background(), 1, PublicOrderInput{}); err == nil {
		t.Error("expected error for empty items")
	}
}

// ─── HandlePaddleTransactionCompleted guard ───────────────────────────────────

func TestHandlePaddleTransactionCompleted_RejectsEmptyTxID(t *testing.T) {
	s := &Service{}
	if err := s.HandlePaddleTransactionCompleted(context.Background(), "", "transaction.completed"); err == nil {
		t.Error("expected error for empty transaction id")
	}
}

// ─── ApplyDiscount guards ─────────────────────────────────────────────────────

func TestApplyDiscount_RejectsZeroOrderID(t *testing.T) {
	s := &Service{}
	if err := s.ApplyDiscount(context.Background(), 1, domain.DiscountInput{DiscountType: "pct", DiscountValue: 10}); err == nil {
		t.Error("expected error for zero order id")
	}
}

func TestApplyDiscount_RejectsBlankType(t *testing.T) {
	s := &Service{}
	if err := s.ApplyDiscount(context.Background(), 1, domain.DiscountInput{OrderID: 1, DiscountValue: 10}); err == nil {
		t.Error("expected error for blank discount type")
	}
}

func TestApplyDiscount_RejectsNegativeValue(t *testing.T) {
	s := &Service{}
	if err := s.ApplyDiscount(context.Background(), 1, domain.DiscountInput{OrderID: 1, DiscountType: "pct", DiscountValue: -5}); err == nil {
		t.Error("expected error for negative discount value")
	}
}
