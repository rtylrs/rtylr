package api

import (
	"testing"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
)

func TestValidateContractTerms(t *testing.T) {
	str := func(s string) *string { return &s }
	cases := []struct {
		name       string
		billing    domain.Billing
		start, end *string
		wantErr    bool
	}{
		{"old MRR-only caller", "", nil, nil, false},
		{"open-ended yearly", domain.BillingYearly, str("2026-11-01"), nil, false},
		{"unknown billing", "weekly", nil, nil, true},
		{"not a date", domain.BillingMonthly, str("next week"), nil, true},
		{"ends before it starts", domain.BillingMonthly, str("2026-11-01"), str("2026-10-31"), true},
	}
	for _, tc := range cases {
		err := validateContractTerms(100, tc.billing, tc.start, tc.end)
		if (err != nil) != tc.wantErr {
			t.Errorf("%s: err = %v, wantErr %v", tc.name, err, tc.wantErr)
		}
	}
	if validateContractTerms(-1, domain.BillingMonthly, nil, nil) == nil {
		t.Error("a negative value must be rejected")
	}
}
