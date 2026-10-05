package domain

import "testing"

func TestTierForPoints_Platinum(t *testing.T) {
	cfg := LoyaltyConfig{MinRedemptionPoints: 100}
	// MinRedemptionPoints * 10 = 1000, so 1000+ should be platinum
	got := cfg.TierFor(1000)
	if got != TierPlatinum {
		t.Errorf("TierFor(1000) = %q, want platinum", got)
	}
}

func TestTierForPoints_BronzeWhenBelowPlatinum(t *testing.T) {
	cfg := LoyaltyConfig{MinRedemptionPoints: 100}
	got := cfg.TierFor(500)
	if got != TierBronze {
		t.Errorf("TierFor(500) = %q, want bronze (below platinum threshold)", got)
	}
}

func TestTierForPoints_NegativeReturnsBronze(t *testing.T) {
	cfg := LoyaltyConfig{MinRedemptionPoints: 100}
	got := cfg.TierFor(-1)
	if got != TierBronze {
		t.Errorf("TierFor(-1) = %q, want bronze", got)
	}
}

func TestCanRedeem_ExactMinimum(t *testing.T) {
	cfg := LoyaltyConfig{MinRedemptionPoints: 200}
	if !cfg.CanRedeem(200) {
		t.Error("expected CanRedeem=true when points == minimum")
	}
}

func TestInitials_SingleCapitalLetter(t *testing.T) {
	m := LoyaltyMember{Name: "Alice"}
	if got := m.Initials(); got != "A" {
		t.Errorf("Initials = %q, want A", got)
	}
}

func TestInitials_DoubleInitials(t *testing.T) {
	m := LoyaltyMember{Name: "Bob Smith"}
	if got := m.Initials(); got != "BS" {
		t.Errorf("Initials = %q, want BS", got)
	}
}

func TestIsInStock_HighStock(t *testing.T) {
	n := 999
	r := LoyaltyReward{Stock: &n}
	if !r.IsInStock() {
		t.Error("expected IsInStock=true for high stock count")
	}
}

func TestIsActive_ReturnsFalseByDefault(t *testing.T) {
	r := LoyaltyReward{}
	if r.IsActive() {
		t.Error("expected IsActive=false when Active field is zero value")
	}
}

func TestTransactionTypeExpiry_IsDistinct(t *testing.T) {
	if TransactionTypeExpiry == TransactionTypeAward {
		t.Error("expiry and award types should be distinct")
	}
	if TransactionTypeExpiry == TransactionTypeRedeem {
		t.Error("expiry and redeem types should be distinct")
	}
}

func TestTransactionTypeManual_IsDistinct(t *testing.T) {
	if TransactionTypeManual == TransactionTypeAward {
		t.Error("manual and award types should be distinct")
	}
	if TransactionTypeManual == TransactionTypeRedeem {
		t.Error("manual and redeem types should be distinct")
	}
}
