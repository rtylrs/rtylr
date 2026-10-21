package infrastructure

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos/orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/orders/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/money"
)

// MarkOrderRefunded was an unconditional
// `SET status='refunded', payment_status='refunded'`, so a $0.01 refund on a
// $100 sale flipped the order out of every revenue query (all of which filter
// payment_status = 'paid') and the whole sale left the books while pos_refunds
// recorded a penny. These cover the decision that replaces it.

func refundRow(total, amountPaid, changeDue string) sqlc.GetOrderForRefundRow {
	row := sqlc.GetOrderForRefundRow{}
	if total != "" {
		row.Total = sql.NullString{String: total, Valid: true}
	}
	if amountPaid != "" {
		row.AmountPaid = sql.NullString{String: amountPaid, Valid: true}
	}
	if changeDue != "" {
		row.ChangeDue = sql.NullString{String: changeDue, Valid: true}
	}
	return row
}

func TestRefundCapCents(t *testing.T) {
	tests := []struct {
		name       string
		total      string
		amountPaid string
		changeDue  string
		wantCents  money.Amount
		wantErr    bool
	}{
		{
			name:       "card sale refunds its total",
			total:      "100.00",
			amountPaid: "100.00",
			wantCents:  10000,
		},
		{
			// The customer handed over 120 and got 20 back, so only 100 ever
			// reached the drawer. Capping at amount_paid would pay out the 20
			// of change the merchant never had.
			name:       "cash sale with change caps at the net kept, not the tender",
			total:      "100.00",
			amountPaid: "120.00",
			changeDue:  "20.00",
			wantCents:  10000,
		},
		{
			// pos_orders.tip_amount is never written by any query, so a
			// tip-inclusive tender is visible only in amount_paid. Capping at
			// total made the tip permanently unrefundable.
			name:       "tip-inclusive tender keeps the tip refundable",
			total:      "100.00",
			amountPaid: "115.00",
			wantCents:  11500,
		},
		{
			name:       "tip plus change caps at the net kept",
			total:      "100.00",
			amountPaid: "120.00",
			changeDue:  "5.00",
			wantCents:  11500,
		},
		{
			// MarkOrderPaidByPaddleTransaction sets payment_status without
			// amount_paid, so amount_paid stays at its column default of 0.
			// Without the total floor such an order could never be refunded.
			name:      "settled without a cash amount still refunds its total",
			total:     "42.00",
			wantCents: 4200,
		},
		{
			name:    "empty total with no tender has nothing to refund",
			wantErr: true,
			total:   "",
		},
		{
			name:    "a zero total has nothing to refund",
			total:   "0.00",
			wantErr: true,
		},
		{
			// An unreadable money column must fail closed, never report a cap
			// of zero-and-therefore-accept-anything.
			name:    "unparseable total is an error",
			total:   "not-a-number",
			wantErr: true,
		},
		{
			name:       "unparseable amount_paid is an error",
			total:      "10.00",
			amountPaid: "abc",
			wantErr:    true,
		},
		{
			name:       "unparseable change_due is an error",
			total:      "10.00",
			amountPaid: "10.00",
			changeDue:  "xx",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := refundCapCents(refundRow(tt.total, tt.amountPaid, tt.changeDue))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("refundCapCents(%q, %q, %q) = %v, want an error",
						tt.total, tt.amountPaid, tt.changeDue, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("refundCapCents(%q, %q, %q): %v", tt.total, tt.amountPaid, tt.changeDue, err)
			}
			if got != tt.wantCents {
				t.Fatalf("refundCapCents(%q, %q, %q) = %d cents, want %d",
					tt.total, tt.amountPaid, tt.changeDue, got, tt.wantCents)
			}
		})
	}
}

func TestSumRefundCents(t *testing.T) {
	tests := []struct {
		name      string
		sum       any
		wantCents money.Amount
		wantErr   bool
	}{
		{name: "no refunds yet", sum: nil, wantCents: 0},
		{name: "nil sum row", sum: []byte(nil), wantCents: 0},
		// MySQL hands a DECIMAL SUM back as []byte on the binary protocol and
		// as a string over the text protocol; both must parse, not be summed
		// as floats.
		{name: "bytes decimal", sum: []byte("12.34"), wantCents: 1234},
		{name: "string decimal", sum: "12.34", wantCents: 1234},
		{name: "float decimal", sum: 12.34, wantCents: 1234},
		{name: "integer cents column", sum: int64(1234), wantCents: 123400},
		{name: "zero sum", sum: []byte("0.00"), wantCents: 0},
		{name: "unparseable sum is an error", sum: []byte("nope"), wantErr: true},
		{name: "unsupported type is an error", sum: struct{}{}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sumRefundCents(tt.sum)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("sumRefundCents(%#v) = %v, want an error", tt.sum, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("sumRefundCents(%#v): %v", tt.sum, err)
			}
			if got != tt.wantCents {
				t.Fatalf("sumRefundCents(%#v) = %d cents, want %d", tt.sum, got, tt.wantCents)
			}
		})
	}
}

func TestRefundPaymentStatusFor(t *testing.T) {
	tests := []struct {
		name            string
		totalRefunded   money.Amount
		capCents        money.Amount
		want            domain.PaymentStatus
		wantKeepsInPaid bool
	}{
		{
			name:            "one cent of a hundred is not a refund",
			totalRefunded:   1,
			capCents:        10000,
			want:            domain.PaymentStatusPaid,
			wantKeepsInPaid: true,
		},
		{
			name:            "a partial refund keeps the sale in the books",
			totalRefunded:   5000,
			capCents:        10000,
			want:            domain.PaymentStatusPaid,
			wantKeepsInPaid: true,
		},
		{
			name:          "reaching the cap retires the sale",
			totalRefunded: 10000,
			capCents:      10000,
			want:          domain.PaymentStatusRefunded,
		},
		{
			// 0.03 + 0.27 is 0.30000000000000004 in float64, which is greater
			// than 0.30, so the last cent of a legitimate full refund used to
			// be refused with a 400.
			name:          "a full refund in awkward cents retires the sale",
			totalRefunded: 3 + 27,
			capCents:      30,
			want:          domain.PaymentStatusRefunded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := refundPaymentStatusFor(tt.totalRefunded, tt.capCents)
			if got != tt.want {
				t.Fatalf("refundPaymentStatusFor(%d, %d) = %q, want %q",
					tt.totalRefunded, tt.capCents, got, tt.want)
			}
			// The enum's 'partial' value would drop the order out of every
			// revenue query exactly as 'refunded' does, so it must never be the
			// answer below the cap.
			if tt.wantKeepsInPaid && got == domain.PaymentStatusPartial {
				t.Fatalf("a sub-cap refund must not be written as 'partial': every revenue query filters payment_status = 'paid'")
			}
		})
	}
}

// The pre-fix over-refund guard was
//
//	interfaceToFloat(prevRefunds) + amount > sqlutil.DecimalToFloat(order.Total)
//
// in float64, which fails closed on a legitimate full refund. This pins both
// halves of that defect so a return to float arithmetic is visible here rather
// than only as a 400 in production.
func TestFloatRefundGuardRefusedValidRefunds(t *testing.T) {
	tests := []struct {
		name          string
		prev          float64
		amount        float64
		cap           float64
		wantFloatAllw bool
		wantCentsOK   bool
	}{
		{
			name:   "thirty cents in two awkward parts",
			prev:   0.03,
			amount: 0.27,
			cap:    0.30,
			// float64: 0.03 + 0.27 == 0.30000000000000004 > 0.30, refused.
			wantFloatAllw: false,
			wantCentsOK:   true,
		},
		{
			name:   "a dollar in awkward parts",
			prev:   0.10,
			amount: 0.20,
			cap:    0.30,
			// 0.1 + 0.2 == 0.30000000000000004 in binary floating point.
			wantFloatAllw: false,
			wantCentsOK:   true,
		},
		{
			name:          "a genuine over-refund is refused by both",
			prev:          0.03,
			amount:        0.28,
			cap:           0.30,
			wantFloatAllw: false,
			wantCentsOK:   false,
		},
		{
			name:          "a plain sub-total refund is allowed by both",
			prev:          0.00,
			amount:        1.00,
			cap:           100.00,
			wantFloatAllw: true,
			wantCentsOK:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			floatAllows := !(tt.prev+tt.amount > tt.cap)
			if floatAllows != tt.wantFloatAllw {
				t.Errorf("float guard allowed=%v, want %v (%.20f > %.2f is %v)",
					floatAllows, tt.wantFloatAllw, tt.prev+tt.amount, tt.cap, tt.prev+tt.amount > tt.cap)
			}

			prior := money.Amount(money.DollarsToCents(tt.prev))
			amount := money.Amount(money.DollarsToCents(tt.amount))
			capCents := money.Amount(money.DollarsToCents(tt.cap))
			centsAllow := prior+amount <= capCents
			if centsAllow != tt.wantCentsOK {
				t.Errorf("cents guard allowed=%v, want %v", centsAllow, tt.wantCentsOK)
			}
		})
	}
}

// The cumulative cap can never be exceeded, in whole cents, for a long series
// of partial refunds of the sizes a cashier actually types.
func TestCumulativeRefundsNeverExceedCap(t *testing.T) {
	tests := []struct {
		name      string
		capCents  money.Amount
		refunds   []money.Amount
		wantAllOK bool
	}{
		{
			// A $100 sale returned a penny at a time: every cent must be
			// accepted, and the hundredth must be the last.
			name:      "ten thousand single cents stop exactly at the cap",
			capCents:  10000,
			refunds:   repeatCents(1, 10000),
			wantAllOK: true,
		},
		{
			name:      "one cent past the cap is refused",
			capCents:  10000,
			refunds:   append(repeatCents(1, 10000), 1),
			wantAllOK: false,
		},
		{
			name:      "awkward cents still sum exactly",
			capCents:  30,
			refunds:   []money.Amount{3, 27},
			wantAllOK: true,
		},
		{
			name:      "a tenth of a cent over is refused",
			capCents:  30,
			refunds:   []money.Amount{3, 28},
			wantAllOK: false,
		},
		{
			name:      "a tip-inclusive sale refunds tip and total",
			capCents:  11500,
			refunds:   []money.Amount{10000, 1500},
			wantAllOK: true,
		},
		{
			name:      "refunding past the tip-inclusive cap is refused",
			capCents:  11500,
			refunds:   []money.Amount{10000, 1501},
			wantAllOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prior := money.Amount(0)
			refusedAt := -1
			for i, r := range tt.refunds {
				total := prior + r
				if total > tt.capCents {
					refusedAt = i
					break
				}
				prior = total
			}
			if tt.wantAllOK {
				if refusedAt >= 0 {
					t.Fatalf("refund %d of %d cents was refused against a %d cent cap",
						refusedAt+1, tt.refunds[refusedAt], tt.capCents)
				}
				if prior != tt.capCents {
					t.Fatalf("accepted refunds totalled %d cents, want the full %d cent cap",
						prior, tt.capCents)
				}
				return
			}
			if refusedAt < 0 {
				t.Fatalf("every refund fit under the %d cent cap, want one refused", tt.capCents)
			}
		})
	}
}

func repeatCents(cents money.Amount, n int) []money.Amount {
	out := make([]money.Amount, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, cents)
	}
	return out
}

// A refund of a voided or cancelled order stays refused, and a paid order
// still refuses a second charge: the states a refund now leaves behind must
// not open either hole.
func TestRefundAndPaymentGuardsHoldAfterAPartialRefund(t *testing.T) {
	// A partially refunded order stays payment_status 'paid' and status 'paid'.
	const status = domain.OrderStatusPaid
	const payment = domain.PaymentStatusPaid

	// The remainder of the sale is still refundable, and the cap is what stops
	// it from being refundable twice over.
	if err := checkRefundable(status, payment); err != nil {
		t.Fatalf("a partially refunded order should still refund, got %v", err)
	}
	// It must not be chargeable again, and it must not be voidable.
	if err := checkPayable(status, payment); !errors.Is(err, domain.ErrOrderAlreadyPaid) {
		t.Fatalf("checkPayable on a settled order = %v, want ErrOrderAlreadyPaid", err)
	}
	if err := checkVoidable(status, payment); !errors.Is(err, domain.ErrOrderAlreadyPaid) {
		t.Fatalf("checkVoidable on a settled order = %v, want ErrOrderAlreadyPaid", err)
	}
	// A voided order is still not refundable, whether or not it took money.
	for _, payment := range []domain.PaymentStatus{domain.PaymentStatusUnpaid, domain.PaymentStatusPaid} {
		if err := checkRefundable(domain.OrderStatusVoided, payment); !errors.Is(err, domain.ErrOrderNotEditable) {
			t.Fatalf("checkRefundable(voided, %q) = %v, want ErrOrderNotEditable", payment, err)
		}
	}
}
