package domain

import "testing"

func TestFullName_TwoParts(t *testing.T) {
	c := Customer{FirstName: "Alice", LastName: "Smith"}
	if got := c.FullName(); got != "Alice Smith" {
		t.Errorf("FullName = %q, want %q", got, "Alice Smith")
	}
}

func TestFullName_FirstOnly(t *testing.T) {
	c := Customer{FirstName: "Alice", LastName: ""}
	if got := c.FullName(); got != "Alice" {
		t.Errorf("FullName = %q, want %q", got, "Alice")
	}
}

func TestFullName_LastOnly(t *testing.T) {
	c := Customer{FirstName: "", LastName: "Smith"}
	if got := c.FullName(); got != "Smith" {
		t.Errorf("FullName = %q, want %q", got, "Smith")
	}
}

func TestHasEmail_WhenSet(t *testing.T) {
	c := Customer{Email: "a@b.com"}
	if !c.HasEmail() {
		t.Error("expected HasEmail=true when email is set")
	}
}

func TestHasEmail_WhenEmpty(t *testing.T) {
	c := Customer{Email: ""}
	if c.HasEmail() {
		t.Error("expected HasEmail=false when email is empty")
	}
}

func TestHasPhone_WhenSet(t *testing.T) {
	c := Customer{Phone: "+1234567890"}
	if !c.HasPhone() {
		t.Error("expected HasPhone=true when phone is set")
	}
}

func TestHasPhone_WhenEmpty(t *testing.T) {
	c := Customer{Phone: ""}
	if c.HasPhone() {
		t.Error("expected HasPhone=false when phone is empty")
	}
}

func TestHasLoyaltyPoints_WhenPositive(t *testing.T) {
	c := Customer{LoyaltyPoints: 100}
	if !c.HasLoyaltyPoints() {
		t.Error("expected HasLoyaltyPoints=true when points > 0")
	}
}

func TestHasLoyaltyPoints_WhenZero(t *testing.T) {
	c := Customer{LoyaltyPoints: 0}
	if c.HasLoyaltyPoints() {
		t.Error("expected HasLoyaltyPoints=false when points = 0")
	}
}

func TestIsHighValue_WhenAboveThreshold(t *testing.T) {
	c := Customer{TotalSpent: 500.0}
	if !c.IsHighValue(250.0) {
		t.Error("expected IsHighValue=true when spent > threshold")
	}
}

func TestIsHighValue_WhenAtThreshold(t *testing.T) {
	c := Customer{TotalSpent: 250.0}
	if !c.IsHighValue(250.0) {
		t.Error("expected IsHighValue=true when spent == threshold")
	}
}

func TestIsHighValue_WhenBelowThreshold(t *testing.T) {
	c := Customer{TotalSpent: 100.0}
	if c.IsHighValue(250.0) {
		t.Error("expected IsHighValue=false when spent < threshold")
	}
}

func TestValidateCustomerInput_Valid(t *testing.T) {
	msg := ValidateCustomerInput(CustomerInput{OwnerID: 1, FirstName: "Alice"})
	if msg != "" {
		t.Errorf("expected no error, got %q", msg)
	}
}

func TestValidateCustomerInput_ZeroOwner(t *testing.T) {
	msg := ValidateCustomerInput(CustomerInput{OwnerID: 0, FirstName: "Alice"})
	if msg == "" {
		t.Error("expected error for zero owner_id")
	}
}

func TestValidateCustomerInput_EmptyName(t *testing.T) {
	msg := ValidateCustomerInput(CustomerInput{OwnerID: 1, FirstName: "", LastName: ""})
	if msg == "" {
		t.Error("expected error when both first and last name are empty")
	}
}

func TestValidateCustomerUpdateInput_Valid(t *testing.T) {
	msg := ValidateCustomerUpdateInput(CustomerUpdateInput{ID: 5, OwnerID: 1})
	if msg != "" {
		t.Errorf("expected no error, got %q", msg)
	}
}

func TestValidateCustomerUpdateInput_ZeroID(t *testing.T) {
	msg := ValidateCustomerUpdateInput(CustomerUpdateInput{ID: 0, OwnerID: 1})
	if msg == "" {
		t.Error("expected error for zero id")
	}
}

func TestValidateAwardPointsInput_Valid(t *testing.T) {
	msg := ValidateAwardPointsInput(AwardPointsInput{OwnerID: 1, CustomerID: 2, Points: 50})
	if msg != "" {
		t.Errorf("expected no error, got %q", msg)
	}
}

func TestValidateAwardPointsInput_ZeroPoints(t *testing.T) {
	msg := ValidateAwardPointsInput(AwardPointsInput{OwnerID: 1, CustomerID: 2, Points: 0})
	if msg == "" {
		t.Error("expected error for zero points")
	}
}

func TestValidateRedeemPointsInput_Valid(t *testing.T) {
	msg := ValidateRedeemPointsInput(RedeemPointsInput{OwnerID: 1, CustomerID: 2, Points: 30})
	if msg != "" {
		t.Errorf("expected no error, got %q", msg)
	}
}

func TestValidateRedeemPointsInput_ZeroCustomer(t *testing.T) {
	msg := ValidateRedeemPointsInput(RedeemPointsInput{OwnerID: 1, CustomerID: 0, Points: 30})
	if msg == "" {
		t.Error("expected error for zero customer_id")
	}
}

func TestIsSentCampaign_WhenSent(t *testing.T) {
	c := Campaign{Status: "sent"}
	if !c.IsSentCampaign() {
		t.Error("expected IsSentCampaign=true when status is sent")
	}
}

func TestIsSentCampaign_WhenDraft(t *testing.T) {
	c := Campaign{Status: "draft"}
	if c.IsSentCampaign() {
		t.Error("expected IsSentCampaign=false when status is draft")
	}
}

func TestIsDraftCampaign_WhenDraft(t *testing.T) {
	c := Campaign{Status: "draft"}
	if !c.IsDraftCampaign() {
		t.Error("expected IsDraftCampaign=true when status is draft")
	}
}

func TestCampaignOpenRate_ZeroSent(t *testing.T) {
	c := Campaign{SentCount: 0, OpenedCount: 0}
	if got := c.CampaignOpenRate(); got != 0 {
		t.Errorf("CampaignOpenRate = %f, want 0 for zero sent count", got)
	}
}

func TestCampaignOpenRate_SomeOpens(t *testing.T) {
	c := Campaign{SentCount: 100, OpenedCount: 40}
	if got := c.CampaignOpenRate(); got != 0.4 {
		t.Errorf("CampaignOpenRate = %f, want 0.4", got)
	}
}

func TestHasAudienceSegment_WhenSet(t *testing.T) {
	id := int64(5)
	c := Campaign{SegmentID: &id}
	if !c.HasAudienceSegment() {
		t.Error("expected HasAudienceSegment=true when segment is set")
	}
}

func TestHasAudienceSegment_WhenNil(t *testing.T) {
	c := Campaign{SegmentID: nil}
	if c.HasAudienceSegment() {
		t.Error("expected HasAudienceSegment=false when segment is nil")
	}
}

func TestLoyaltyConfigIsEnabled_WhenTrue(t *testing.T) {
	lc := LoyaltyConfig{IsEnabled: true}
	if !lc.LoyaltyConfigIsEnabled() {
		t.Error("expected LoyaltyConfigIsEnabled=true")
	}
}

func TestLoyaltyConfigIsEnabled_WhenFalse(t *testing.T) {
	lc := LoyaltyConfig{IsEnabled: false}
	if lc.LoyaltyConfigIsEnabled() {
		t.Error("expected LoyaltyConfigIsEnabled=false")
	}
}

func TestHasPointsExpiry_WhenPositive(t *testing.T) {
	lc := LoyaltyConfig{PointsExpiryDays: 365}
	if !lc.HasPointsExpiry() {
		t.Error("expected HasPointsExpiry=true when expiry days > 0")
	}
}

func TestHasPointsExpiry_WhenZero(t *testing.T) {
	lc := LoyaltyConfig{PointsExpiryDays: 0}
	if lc.HasPointsExpiry() {
		t.Error("expected HasPointsExpiry=false when expiry days = 0")
	}
}
