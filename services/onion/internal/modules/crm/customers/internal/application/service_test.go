package application

import (
	"testing"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
)

func TestRecalculateTier_WithExplicitConfig(t *testing.T) {
	cfg := domain.LoyaltyConfig{
		SilverMin:   100,
		GoldMin:     500,
		PlatinumMin: 2000,
	}
	cases := []struct {
		points int
		want   string
	}{
		{0, domain.TierBronze},
		{50, domain.TierBronze},
		{99, domain.TierBronze},
		{100, domain.TierSilver},
		{499, domain.TierSilver},
		{500, domain.TierGold},
		{1999, domain.TierGold},
		{2000, domain.TierPlatinum},
		{1_000_000, domain.TierPlatinum},
	}
	for _, tc := range cases {
		if got := RecalculateTier(tc.points, cfg); got != tc.want {
			t.Errorf("RecalculateTier(%d) = %q, want %q", tc.points, got, tc.want)
		}
	}
}

func TestRecalculateTier_DefaultsWhenConfigZero(t *testing.T) {
	// Empty config: defaults are silver=500, gold=1500, platinum=5000.
	cfg := domain.LoyaltyConfig{}
	cases := []struct {
		points int
		want   string
	}{
		{0, domain.TierBronze},
		{499, domain.TierBronze},
		{500, domain.TierSilver},
		{1500, domain.TierGold},
		{5000, domain.TierPlatinum},
	}
	for _, tc := range cases {
		if got := RecalculateTier(tc.points, cfg); got != tc.want {
			t.Errorf("default RecalculateTier(%d) = %q, want %q", tc.points, got, tc.want)
		}
	}
}

func TestLoyaltyPointsFromCents_RoundsCurrencyBeforeRate(t *testing.T) {
	if got := loyaltyPointsFromCents(99.99, 0.1); got != 10 {
		t.Fatalf("loyaltyPointsFromCents(99.99, 0.1) = %d, want 10", got)
	}
}
