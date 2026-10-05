package infrastructure

import (
	"strings"
	"testing"

	"github.com/voxire/onion/internal/platform/money"
)

// The three lines must close exactly, for every gross and net a payroll run
// can produce. This is the assertion #1183 is about: it used to be
// roundMoney(totalDebits-totalCredits) != 0, which snapped the difference to
// two decimals before comparing it to zero and so could not detect an error
// smaller than half a cent.
func TestDerivePayrollAmountsClosesExactly(t *testing.T) {
	cases := []struct {
		name           string
		gross, net     float64
		wantGross      money.Amount
		wantNet        money.Amount
		wantDeductions money.Amount
	}{
		{"whole amounts", 1000, 800, 100000, 80000, 20000},
		{"cents on both sides", 1234.56, 1000.01, 123456, 100001, 23455},
		{"no deductions", 500.25, 500.25, 50025, 50025, 0},
		{"a single cent of deductions", 100.01, 100, 10001, 10000, 1},
		{"zero", 0, 0, 0, 0, 0},
		// 0.07 does not scale cleanly by 100 in float64, which is the shape of
		// value the old path could mis-round.
		{"an amount that scales inexactly", 0.07, 0.03, 7, 3, 4},
		{"tenths, which float64 cannot sum", 0.3, 0.1, 30, 10, 20},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := derivePayrollAmounts(tc.gross, tc.net)
			if err != nil {
				t.Fatalf("derivePayrollAmounts(%v, %v): %v", tc.gross, tc.net, err)
			}
			if got.gross != tc.wantGross || got.net != tc.wantNet || got.deductions != tc.wantDeductions {
				t.Errorf("got gross=%d net=%d deductions=%d, want %d/%d/%d",
					got.gross, got.net, got.deductions, tc.wantGross, tc.wantNet, tc.wantDeductions)
			}
			// The property that matters, stated directly.
			if got.gross != got.net+got.deductions {
				t.Errorf("entry does not close: %d != %d + %d", got.gross, got.net, got.deductions)
			}
		})
	}
}

// Swept across a range of cent values rather than a handful of examples,
// because the old defect was a rounding one: it appeared for particular
// values, not for an obviously wrong input.
func TestDerivePayrollAmountsClosesAcrossASweep(t *testing.T) {
	for grossCents := 0; grossCents <= 2000; grossCents++ {
		for _, netOffset := range []int{0, 1, 7, 33, 99, 100, 337} {
			netCents := grossCents - netOffset
			if netCents < 0 {
				continue
			}
			gross := float64(grossCents) / 100
			net := float64(netCents) / 100

			got, err := derivePayrollAmounts(gross, net)
			if err != nil {
				t.Fatalf("derivePayrollAmounts(%v, %v): %v", gross, net, err)
			}
			if int(got.gross) != grossCents || int(got.net) != netCents {
				t.Fatalf("gross/net round trip failed: got %d/%d, want %d/%d",
					got.gross, got.net, grossCents, netCents)
			}
			if got.gross != got.net+got.deductions {
				t.Fatalf("entry does not close at gross=%d net=%d: %d != %d + %d",
					grossCents, netCents, got.gross, got.net, got.deductions)
			}
			if int(got.deductions) != netOffset {
				t.Fatalf("deductions = %d, want %d", got.deductions, netOffset)
			}
		}
	}
}

// A payroll export carrying a third decimal is refused rather than rounded
// into the ledger. The old path rounded it and the balance check, being made
// of the same rounding, saw nothing wrong.
func TestDerivePayrollAmountsRefusesAmountsThatAreNotWholeCents(t *testing.T) {
	cases := []struct{ gross, net float64 }{
		{9.996, 5},
		{10, 9.996},
		{1.0001, 1},
	}
	for _, tc := range cases {
		if _, err := derivePayrollAmounts(tc.gross, tc.net); err == nil {
			t.Errorf("derivePayrollAmounts(%v, %v) returned no error for a sub-cent amount", tc.gross, tc.net)
		}
	}
}

// Net above gross is not a rounding artefact, it is a bad export, and the
// error should say so rather than mention rounding as the old one did.
func TestDerivePayrollAmountsRefusesNetAboveGross(t *testing.T) {
	_, err := derivePayrollAmounts(100, 150.25)
	if err == nil {
		t.Fatal("net above gross was accepted")
	}
	if !strings.Contains(err.Error(), "negative") {
		t.Errorf("error = %q, want it to name the negative deductions", err)
	}
	// The amounts belong in the message, exactly, so an operator can see them.
	if !strings.Contains(err.Error(), "100.00") || !strings.Contains(err.Error(), "150.25") {
		t.Errorf("error = %q, want the exact gross and net in it", err)
	}
}

// What reaches the DECIMAL(15,2) column is a decimal string, not a float, so
// the driver never formats an amount.
func TestPayrollLinesRenderExactDecimals(t *testing.T) {
	amounts, err := derivePayrollAmounts(1234.56, 1000.01)
	if err != nil {
		t.Fatalf("derivePayrollAmounts: %v", err)
	}
	lines := []payrollLine{
		{accountID: 1, debit: amounts.gross},
		{accountID: 2, credit: amounts.net},
		{accountID: 3, credit: amounts.deductions},
	}
	want := []string{"1234.56", "1000.01", "234.55"}
	got := []string{lines[0].debit.String(), lines[1].credit.String(), lines[2].credit.String()}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d rendered %q, want %q", i, got[i], want[i])
		}
	}
}
