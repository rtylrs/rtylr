package domain

import "testing"

func TestFullName_BothEmpty(t *testing.T) {
	c := Customer{FirstName: "", LastName: ""}
	if got := c.FullName(); got != "" {
		t.Errorf("FullName = %q, want empty string", got)
	}
}

func TestFullName_HasNoLeadingOrTrailingSpace(t *testing.T) {
	c := Customer{FirstName: "Alice", LastName: ""}
	if got := c.FullName(); got != "Alice" {
		t.Errorf("FullName = %q, want %q", got, "Alice")
	}
}

func TestIsHighValue_ExactThreshold(t *testing.T) {
	c := Customer{TotalSpent: 1000.0}
	if !c.IsHighValue(1000.0) {
		t.Error("expected IsHighValue=true at exact threshold")
	}
}

func TestValidateCustomerInput_LastNameOnlyIsValid(t *testing.T) {
	msg := ValidateCustomerInput(CustomerInput{OwnerID: 1, FirstName: "", LastName: "Smith"})
	if msg != "" {
		t.Errorf("expected no error when last name only provided, got %q", msg)
	}
}

func TestValidateCustomerUpdateInput_ZeroOwner(t *testing.T) {
	msg := ValidateCustomerUpdateInput(CustomerUpdateInput{ID: 5, OwnerID: 0})
	if msg == "" {
		t.Error("expected error for zero owner_id")
	}
}

func TestCampaignOpenRate_HighOpenCount(t *testing.T) {
	c := Campaign{SentCount: 10, OpenedCount: 8}
	if got := c.CampaignOpenRate(); got != 0.8 {
		t.Errorf("CampaignOpenRate = %f, want 0.8", got)
	}
}

func TestIsDraftCampaign_WhenSent(t *testing.T) {
	c := Campaign{Status: "sent"}
	if c.IsDraftCampaign() {
		t.Error("expected IsDraftCampaign=false for sent campaign")
	}
}

func TestLoyaltyConfigIsEnabled_Toggles(t *testing.T) {
	enabled := LoyaltyConfig{IsEnabled: true}
	disabled := LoyaltyConfig{IsEnabled: false}
	if !enabled.LoyaltyConfigIsEnabled() {
		t.Error("expected LoyaltyConfigIsEnabled=true")
	}
	if disabled.LoyaltyConfigIsEnabled() {
		t.Error("expected LoyaltyConfigIsEnabled=false")
	}
}

func TestHasPointsExpiry_NegativeDays(t *testing.T) {
	lc := LoyaltyConfig{PointsExpiryDays: -1}
	if lc.HasPointsExpiry() {
		t.Error("expected HasPointsExpiry=false for negative expiry days")
	}
}

func TestTierConstants_AreFour(t *testing.T) {
	tiers := []string{TierBronze, TierSilver, TierGold, TierPlatinum}
	if len(tiers) != 4 {
		t.Errorf("expected 4 tier constants, got %d", len(tiers))
	}
}

func TestValidateAwardPointsInput_NegativePoints(t *testing.T) {
	msg := ValidateAwardPointsInput(AwardPointsInput{OwnerID: 1, CustomerID: 2, Points: -5})
	if msg == "" {
		t.Error("expected error for negative points")
	}
}

func TestValidateRedeemPointsInput_NegativePoints(t *testing.T) {
	msg := ValidateRedeemPointsInput(RedeemPointsInput{OwnerID: 1, CustomerID: 2, Points: -1})
	if msg == "" {
		t.Error("expected error for negative points")
	}
}
