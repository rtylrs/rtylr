package infrastructure

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
)

// The order state machine, exercised over every (status, payment_status) pair
// the module can produce. A void used to be an unconditional status write, so a
// paid or refunded order could be relabelled voided and then paid out again;
// these are the transitions the repositories now enforce before writing.

type orderState struct {
	status        domain.OrderStatus
	paymentStatus domain.PaymentStatus
}

func TestCheckVoidable(t *testing.T) {
	tests := []struct {
		name  string
		state orderState
		want  error
	}{
		{"pending order voids", orderState{domain.OrderStatusPending, domain.PaymentStatusUnpaid}, nil},
		{"confirmed order voids", orderState{domain.OrderStatusConfirmed, domain.PaymentStatusUnpaid}, nil},
		{"preparing order voids", orderState{domain.OrderStatusPreparing, domain.PaymentStatusUnpaid}, nil},
		{"ready order voids", orderState{domain.OrderStatusReady, domain.PaymentStatusUnpaid}, nil},
		{"served order voids", orderState{domain.OrderStatusServed, domain.PaymentStatusUnpaid}, nil},
		{"held order voids", orderState{domain.OrderStatusHeld, domain.PaymentStatusUnpaid}, nil},
		{"paid order is refused", orderState{domain.OrderStatusPaid, domain.PaymentStatusPaid}, domain.ErrOrderAlreadyPaid},
		{"deposit recorded as partial is refused", orderState{domain.OrderStatusPending, domain.PaymentStatusPartial}, domain.ErrOrderAlreadyPaid},
		{"refunded order is refused", orderState{domain.OrderStatusRefunded, domain.PaymentStatusRefunded}, domain.ErrOrderAlreadyPaid},
		// A voided row whose payment_status still reads paid is the shape the
		// unguarded void produced. The money check must win here: refunding or
		// re-charging it is what double-pays the merchant.
		{"voided order that took money is refused", orderState{domain.OrderStatusVoided, domain.PaymentStatusPaid}, domain.ErrOrderAlreadyPaid},
		{"cancelled order is refused", orderState{domain.OrderStatusCancelled, domain.PaymentStatusUnpaid}, domain.ErrOrderNotEditable},
		{"re-sent void is idempotent", orderState{domain.OrderStatusVoided, domain.PaymentStatusUnpaid}, errOrderAlreadyVoided},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkVoidable(tt.state.status, tt.state.paymentStatus)
			if tt.want == nil {
				if got != nil {
					t.Fatalf("checkVoidable(%q, %q) = %v, want nil", tt.state.status, tt.state.paymentStatus, got)
				}
				return
			}
			if !errors.Is(got, tt.want) {
				t.Fatalf("checkVoidable(%q, %q) = %v, want %v", tt.state.status, tt.state.paymentStatus, got, tt.want)
			}
		})
	}
}

func TestCheckPayable(t *testing.T) {
	tests := []struct {
		name  string
		state orderState
		want  error
	}{
		{"pending order takes payment", orderState{domain.OrderStatusPending, domain.PaymentStatusUnpaid}, nil},
		{"served order takes payment", orderState{domain.OrderStatusServed, domain.PaymentStatusUnpaid}, nil},
		{"held order takes payment", orderState{domain.OrderStatusHeld, domain.PaymentStatusUnpaid}, nil},
		{"deposit can be paid off", orderState{domain.OrderStatusPending, domain.PaymentStatusPartial}, nil},
		{"paid order is refused", orderState{domain.OrderStatusPaid, domain.PaymentStatusPaid}, domain.ErrOrderAlreadyPaid},
		{"refunded order is refused", orderState{domain.OrderStatusRefunded, domain.PaymentStatusRefunded}, domain.ErrOrderNotEditable},
		{"voided order is refused", orderState{domain.OrderStatusVoided, domain.PaymentStatusUnpaid}, domain.ErrOrderNotEditable},
		{"cancelled order is refused", orderState{domain.OrderStatusCancelled, domain.PaymentStatusUnpaid}, domain.ErrOrderNotEditable},
		// Charged after a refund: the customer pays twice and the shift close
		// still subtracts the earlier refund from expected cash.
		{"refunded payment status is refused whatever the workflow says", orderState{domain.OrderStatusPending, domain.PaymentStatusRefunded}, domain.ErrOrderNotEditable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkPayable(tt.state.status, tt.state.paymentStatus)
			if tt.want == nil {
				if got != nil {
					t.Fatalf("checkPayable(%q, %q) = %v, want nil", tt.state.status, tt.state.paymentStatus, got)
				}
				return
			}
			if !errors.Is(got, tt.want) {
				t.Fatalf("checkPayable(%q, %q) = %v, want %v", tt.state.status, tt.state.paymentStatus, got, tt.want)
			}
		})
	}
}

func TestCheckRefundable(t *testing.T) {
	tests := []struct {
		name  string
		state orderState
		want  error
	}{
		{"paid order refunds", orderState{domain.OrderStatusPaid, domain.PaymentStatusPaid}, nil},
		{"partially refunded order still refunds", orderState{domain.OrderStatusRefunded, domain.PaymentStatusRefunded}, nil},
		{"unpaid order does not refund", orderState{domain.OrderStatusPending, domain.PaymentStatusUnpaid}, domain.ErrOrderNotPaid},
		{"voided order does not refund", orderState{domain.OrderStatusVoided, domain.PaymentStatusUnpaid}, domain.ErrOrderNotEditable},
		{"cancelled order does not refund", orderState{domain.OrderStatusCancelled, domain.PaymentStatusUnpaid}, domain.ErrOrderNotEditable},
		// Void after payment left payment_status at paid; refunding such a row
		// pays the customer back for a sale the record says never happened.
		{"voided order that took money does not refund", orderState{domain.OrderStatusVoided, domain.PaymentStatusPaid}, domain.ErrOrderNotEditable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkRefundable(tt.state.status, tt.state.paymentStatus)
			if tt.want == nil {
				if got != nil {
					t.Fatalf("checkRefundable(%q, %q) = %v, want nil", tt.state.status, tt.state.paymentStatus, got)
				}
				return
			}
			if !errors.Is(got, tt.want) {
				t.Fatalf("checkRefundable(%q, %q) = %v, want %v", tt.state.status, tt.state.paymentStatus, got, tt.want)
			}
		})
	}
}

// A void may never be the transition that first records money against an order,
// and a refund may never be the transition that first erases a settled sale.
func TestVoidRefundAndPaymentCannotCrossMoneyStates(t *testing.T) {
	if err := checkVoidable(domain.OrderStatusPaid, domain.PaymentStatusPaid); err == nil {
		t.Fatal("paid order was voidable")
	}
	if err := checkRefundable(domain.OrderStatusVoided, domain.PaymentStatusPaid); err == nil {
		t.Fatal("voided order was refundable")
	}
	if err := checkPayable(domain.OrderStatusRefunded, domain.PaymentStatusRefunded); err == nil {
		t.Fatal("refunded order was payable")
	}
}
