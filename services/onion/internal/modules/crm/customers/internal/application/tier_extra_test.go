package application

import (
	"testing"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
)

func TestRecalculateTier_BronzeAtZero(t *testing.T) {
	cfg := domain.LoyaltyConfig{SilverMin: 100, GoldMin: 500, PlatinumMin: 2000}
	if got := RecalculateTier(0, cfg); got != domain.TierBronze {
		t.Errorf("expected bronze at 0 points, got %q", got)
	}
}

func TestRecalculateTier_SilverAtThreshold(t *testing.T) {
	cfg := domain.LoyaltyConfig{SilverMin: 200, GoldMin: 600, PlatinumMin: 1200}
	if got := RecalculateTier(200, cfg); got != domain.TierSilver {
		t.Errorf("expected silver at 200 points, got %q", got)
	}
}

func TestRecalculateTier_GoldAtThreshold(t *testing.T) {
	cfg := domain.LoyaltyConfig{SilverMin: 200, GoldMin: 600, PlatinumMin: 1200}
	if got := RecalculateTier(600, cfg); got != domain.TierGold {
		t.Errorf("expected gold at 600 points, got %q", got)
	}
}

func TestRecalculateTier_PlatinumAtThreshold(t *testing.T) {
	cfg := domain.LoyaltyConfig{SilverMin: 200, GoldMin: 600, PlatinumMin: 1200}
	if got := RecalculateTier(1200, cfg); got != domain.TierPlatinum {
		t.Errorf("expected platinum at 1200 points, got %q", got)
	}
}

func TestRecalculateTier_JustBelowSilver(t *testing.T) {
	cfg := domain.LoyaltyConfig{SilverMin: 500, GoldMin: 1500, PlatinumMin: 5000}
	if got := RecalculateTier(499, cfg); got != domain.TierBronze {
		t.Errorf("expected bronze at 499 points, got %q", got)
	}
}

func TestRecalculateTier_JustBelowGold(t *testing.T) {
	cfg := domain.LoyaltyConfig{SilverMin: 500, GoldMin: 1500, PlatinumMin: 5000}
	if got := RecalculateTier(1499, cfg); got != domain.TierSilver {
		t.Errorf("expected silver at 1499 points, got %q", got)
	}
}

func TestRecalculateTier_JustBelowPlatinum(t *testing.T) {
	cfg := domain.LoyaltyConfig{SilverMin: 500, GoldMin: 1500, PlatinumMin: 5000}
	if got := RecalculateTier(4999, cfg); got != domain.TierGold {
		t.Errorf("expected gold at 4999 points, got %q", got)
	}
}

func TestRecalculateTier_VeryHighPoints(t *testing.T) {
	cfg := domain.LoyaltyConfig{SilverMin: 100, GoldMin: 500, PlatinumMin: 2000}
	if got := RecalculateTier(999999, cfg); got != domain.TierPlatinum {
		t.Errorf("expected platinum at 999999 points, got %q", got)
	}
}

func TestRecalculateTier_NegativePoints_ReturnsBronze(t *testing.T) {
	cfg := domain.LoyaltyConfig{SilverMin: 100, GoldMin: 500, PlatinumMin: 2000}
	if got := RecalculateTier(-10, cfg); got != domain.TierBronze {
		t.Errorf("expected bronze for negative points, got %q", got)
	}
}
