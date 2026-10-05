package infrastructure

import (
	"testing"

	"github.com/voxire/onion/internal/modules/pos/orders/internal/domain"
)

// pos_orders.payment_method is ENUM('cash','card','online','split'). It has no
// empty member, so an order that has not been paid yet must be written as NULL.
//
// SplitOrder marked the copied value Valid unconditionally, so splitting an
// unpaid order sent ” and MySQL rejected the row with "Error 1265: Data
// truncated for column 'payment_method'" (#1185). This is that decision, now in
// one place and testable without a database.
func TestNullPaymentMethodWritesNullForAnUnsetMethod(t *testing.T) {
	got := nullPaymentMethod("")
	if got.Valid {
		t.Fatal("an unset payment method was marked valid; MySQL rejects '' for this ENUM")
	}
}

// Every member the column accepts must survive, or a paid order loses how it
// was paid.
func TestNullPaymentMethodKeepsEveryEnumMember(t *testing.T) {
	members := []domain.PaymentMethod{
		domain.PaymentMethodCash,
		domain.PaymentMethodCard,
		domain.PaymentMethodOnline,
		domain.PaymentMethodSplit,
	}
	for _, method := range members {
		got := nullPaymentMethod(method)
		if !got.Valid {
			t.Errorf("%q was marked invalid", method)
			continue
		}
		if string(got.PosOrdersPaymentMethod) != string(method) {
			t.Errorf("%q was written as %q", method, got.PosOrdersPaymentMethod)
		}
	}
}

// A typo is passed through rather than nulled, so the column rejects it and the
// mistake stays visible. Nulling it would silently drop what the caller asked
// for.
func TestNullPaymentMethodPassesAnUnknownValueThrough(t *testing.T) {
	got := nullPaymentMethod("bank_transfer")
	if !got.Valid {
		t.Fatal("an unknown method was nulled, which hides the caller's mistake")
	}
	if string(got.PosOrdersPaymentMethod) != "bank_transfer" {
		t.Errorf("value = %q, want it passed through unchanged", got.PosOrdersPaymentMethod)
	}
}
