package domain

import (
	"testing"

	"github.com/voxire/onion/internal/platform/money"
)

func TestMonthlyEquivalent(t *testing.T) {
	cases := []struct {
		billing Billing
		want    money.Amount
	}{
		{BillingMonthly, 120000},
		{BillingYearly, 10000},
		{BillingOneTime, 0},
	}
	for _, tc := range cases {
		if got := MonthlyEquivalent(120000, tc.billing); got != tc.want {
			t.Errorf("MonthlyEquivalent(1200.00, %s) = %d, want %d", tc.billing, got, tc.want)
		}
	}
}
