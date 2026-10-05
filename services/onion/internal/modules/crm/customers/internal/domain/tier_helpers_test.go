package domain

import "testing"

func TestValidateLoyaltyTierInput_Valid(t *testing.T) {
	msg := ValidateLoyaltyTierInput(LoyaltyTierInput{OwnerID: 1, Name: "Gold"})
	if msg != "" {
		t.Errorf("expected no error, got %q", msg)
	}
}

func TestValidateLoyaltyTierInput_ZeroOwner(t *testing.T) {
	msg := ValidateLoyaltyTierInput(LoyaltyTierInput{OwnerID: 0, Name: "Gold"})
	if msg == "" {
		t.Error("expected error for zero owner_id")
	}
}

func TestValidateLoyaltyTierInput_EmptyName(t *testing.T) {
	msg := ValidateLoyaltyTierInput(LoyaltyTierInput{OwnerID: 1, Name: ""})
	if msg == "" {
		t.Error("expected error for empty name")
	}
}

func TestHasPerks_WhenSet(t *testing.T) {
	lt := LoyaltyTier{Perks: []byte(`[{"type":"free_shipping"}]`)}
	if !lt.HasPerks() {
		t.Error("expected HasPerks=true when perks array is non-empty")
	}
}

func TestHasPerks_WhenEmptyArray(t *testing.T) {
	lt := LoyaltyTier{Perks: []byte(`[]`)}
	if lt.HasPerks() {
		t.Error("expected HasPerks=false when perks is empty array")
	}
}

func TestHasPerks_WhenNull(t *testing.T) {
	lt := LoyaltyTier{Perks: []byte(`null`)}
	if lt.HasPerks() {
		t.Error("expected HasPerks=false when perks is null")
	}
}

func TestHasPerks_WhenNil(t *testing.T) {
	lt := LoyaltyTier{Perks: nil}
	if lt.HasPerks() {
		t.Error("expected HasPerks=false when perks is nil")
	}
}

func TestHasDiscount_WhenPositive(t *testing.T) {
	lt := LoyaltyTier{DiscountPct: 10.0}
	if !lt.HasDiscount() {
		t.Error("expected HasDiscount=true when discount > 0")
	}
}

func TestHasDiscount_WhenZero(t *testing.T) {
	lt := LoyaltyTier{DiscountPct: 0}
	if lt.HasDiscount() {
		t.Error("expected HasDiscount=false when discount = 0")
	}
}

func TestIsEarned_WhenPointsAboveMinimum(t *testing.T) {
	lt := LoyaltyTier{MinPoints: 500}
	if !lt.IsEarned(600) {
		t.Error("expected IsEarned=true when points >= min_points")
	}
}

func TestIsEarned_WhenPointsAtMinimum(t *testing.T) {
	lt := LoyaltyTier{MinPoints: 500}
	if !lt.IsEarned(500) {
		t.Error("expected IsEarned=true when points == min_points")
	}
}

func TestIsEarned_WhenPointsBelowMinimum(t *testing.T) {
	lt := LoyaltyTier{MinPoints: 500}
	if lt.IsEarned(499) {
		t.Error("expected IsEarned=false when points < min_points")
	}
}

func TestIsEarned_WhenMinimumIsZero(t *testing.T) {
	lt := LoyaltyTier{MinPoints: 0}
	if !lt.IsEarned(0) {
		t.Error("expected IsEarned=true when min_points=0 and points=0")
	}
}
