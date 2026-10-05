package domain

import "testing"

func TestValidateLoyaltyConfigInput_Valid(t *testing.T) {
	msg := ValidateLoyaltyConfigInput(LoyaltyConfigInput{OwnerID: 1})
	if msg != "" {
		t.Errorf("expected no error, got %q", msg)
	}
}

func TestValidateLoyaltyConfigInput_ZeroOwner(t *testing.T) {
	msg := ValidateLoyaltyConfigInput(LoyaltyConfigInput{OwnerID: 0})
	if msg == "" {
		t.Error("expected error for zero owner_id")
	}
}

func TestCanRedeem_WhenAboveMinimum(t *testing.T) {
	cfg := LoyaltyConfig{MinRedemptionPoints: 100}
	if !cfg.CanRedeem(150) {
		t.Error("expected CanRedeem=true when points >= minimum")
	}
}

func TestCanRedeem_WhenBelowMinimum(t *testing.T) {
	cfg := LoyaltyConfig{MinRedemptionPoints: 100}
	if cfg.CanRedeem(50) {
		t.Error("expected CanRedeem=false when points < minimum")
	}
}

func TestCanRedeem_WhenMinimumIsZero(t *testing.T) {
	cfg := LoyaltyConfig{MinRedemptionPoints: 0}
	if !cfg.CanRedeem(1) {
		t.Error("expected CanRedeem=true when minimum is 0 and points > 0")
	}
}

func TestCanRedeem_WhenMinimumIsZeroAndPointsZero(t *testing.T) {
	cfg := LoyaltyConfig{MinRedemptionPoints: 0}
	if cfg.CanRedeem(0) {
		t.Error("expected CanRedeem=false when minimum is 0 and points = 0")
	}
}

func TestHasReferenceID_WhenSet(t *testing.T) {
	id := int64(42)
	tx := LoyaltyTransaction{ReferenceID: &id}
	if !tx.HasReferenceID() {
		t.Error("expected HasReferenceID=true when reference_id is set")
	}
}

func TestHasReferenceID_WhenNil(t *testing.T) {
	tx := LoyaltyTransaction{ReferenceID: nil}
	if tx.HasReferenceID() {
		t.Error("expected HasReferenceID=false when reference_id is nil")
	}
}

func TestTierConstants_AreFourDistinct(t *testing.T) {
	tiers := []string{TierBronze, TierSilver, TierGold, TierPlatinum}
	if len(tiers) != 4 {
		t.Errorf("expected 4 tier constants, got %d", len(tiers))
	}
	seen := make(map[string]bool)
	for _, tier := range tiers {
		if seen[tier] {
			t.Errorf("duplicate tier: %q", tier)
		}
		seen[tier] = true
	}
}

func TestIsAward_AwardTypeString(t *testing.T) {
	if TransactionTypeAward != "award" {
		t.Errorf("TransactionTypeAward = %q, want %q", TransactionTypeAward, "award")
	}
}

func TestIsRedeem_RedeemTypeString(t *testing.T) {
	if TransactionTypeRedeem != "redeem" {
		t.Errorf("TransactionTypeRedeem = %q, want %q", TransactionTypeRedeem, "redeem")
	}
}
