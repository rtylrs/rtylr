package application

import "testing"

func TestBudgetTier(t *testing.T) {
	cases := []struct {
		budget string
		want   string
	}{
		// Contact form bands
		{"Under $2K", "NEW"},
		{"$2K - $8K", "WARM"},
		{"$8K - $20K", "HOT"},
		{"$20K+", "HOT"},
		{"Let's figure it out", "NEW"},
		// Quote form bands
		{"Under $2k", "NEW"},
		{"$2k-$5k", "WARM"},
		{"$5k-$15k", "HOT"},
		{"$15k+", "HOT"},
		{"Let's discuss", "NEW"},
		// Free text and edge cases
		{"", "NEW"},
		{"around 10,000 USD", "WARM"},
		{"25,000", "HOT"},
		{"$3k to $6k", "WARM"},
		{"no budget yet", "NEW"},
		{"4k", "NEW"},
		{"71234567", "NEW"},
		{"20260115", "NEW"},
		{"10000", "NEW"},
		{"call me on 03 940 708", "NEW"},
		{"we need 15 key pages redone", "NEW"},
		{"8 keywords to target", "NEW"},
	}
	for _, tc := range cases {
		t.Run(tc.budget, func(t *testing.T) {
			if got := budgetTier(tc.budget); got != tc.want {
				t.Fatalf("budgetTier(%q) = %q, want %q", tc.budget, got, tc.want)
			}
		})
	}
}
