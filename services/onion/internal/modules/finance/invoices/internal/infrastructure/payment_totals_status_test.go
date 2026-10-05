package infrastructure

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
	"github.com/voxire/onion/internal/platform/money"
)

// RecordPayment is the only writer of the money-derived statuses, so the
// derivation it applies is what decides revenue and the remaining balance.
func TestPaymentTotalsStatus(t *testing.T) {
	paymentDate := time.Date(2026, time.June, 20, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name       string
		dueCents   money.Amount
		wantStatus string
		wantPaid   bool
	}{
		{"balance still outstanding", money.Amount(3666), domain.StatusPartial, false},
		{"one cent short", money.Amount(1), domain.StatusPartial, false},
		{"settled", money.Amount(0), domain.StatusPaid, true},
		{"overpaid is clamped by the caller", money.Amount(0), domain.StatusPaid, true},
	}
	for _, tc := range cases {
		status, paidDate := paymentTotalsStatus(tc.dueCents, paymentDate)
		if status != tc.wantStatus {
			t.Errorf("%s: status = %q, want %q", tc.name, status, tc.wantStatus)
		}
		if (paidDate != nil) != tc.wantPaid {
			t.Errorf("%s: paid date = %v, want set = %t", tc.name, paidDate, tc.wantPaid)
		}
		if paidDate != nil && !paidDate.Equal(paymentDate) {
			t.Errorf("%s: paid date = %v, want %v", tc.name, paidDate, paymentDate)
		}
	}
}
