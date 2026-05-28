package domain

import "testing"

func TestIsAward_WhenAward(t *testing.T) {
	tx := LoyaltyTransaction{Type: TransactionTypeAward}
	if !tx.IsAward() {
		t.Error("expected IsAward=true for award transaction")
	}
}

func TestIsAward_WhenRedeem(t *testing.T) {
	tx := LoyaltyTransaction{Type: TransactionTypeRedeem}
	if tx.IsAward() {
		t.Error("expected IsAward=false for redeem transaction")
	}
}

func TestIsRedeem_WhenRedeem(t *testing.T) {
	tx := LoyaltyTransaction{Type: TransactionTypeRedeem}
	if !tx.IsRedeem() {
		t.Error("expected IsRedeem=true for redeem transaction")
	}
}

func TestIsRedeem_WhenAward(t *testing.T) {
	tx := LoyaltyTransaction{Type: TransactionTypeAward}
	if tx.IsRedeem() {
		t.Error("expected IsRedeem=false for award transaction")
	}
}

func TestInitials_TwoWords(t *testing.T) {
	m := LoyaltyMember{Name: "Alice Smith"}
	if got := m.Initials(); got != "AS" {
		t.Errorf("Initials = %q, want AS", got)
	}
}

func TestInitials_OneWord(t *testing.T) {
	m := LoyaltyMember{Name: "Madonna"}
	if got := m.Initials(); got != "M" {
		t.Errorf("Initials = %q, want M", got)
	}
}

func TestInitials_Empty(t *testing.T) {
	m := LoyaltyMember{Name: ""}
	if got := m.Initials(); got != "" {
		t.Errorf("Initials for empty name = %q, want empty", got)
	}
}

func TestInitials_ThreeWords_TakesFirst2(t *testing.T) {
	m := LoyaltyMember{Name: "John Paul Jones"}
	if got := m.Initials(); got != "JP" {
		t.Errorf("Initials = %q, want JP", got)
	}
}

func TestIsActive_WhenTrue(t *testing.T) {
	r := LoyaltyReward{Active: true}
	if !r.IsActive() {
		t.Error("expected IsActive=true")
	}
}

func TestIsActive_WhenFalse(t *testing.T) {
	r := LoyaltyReward{Active: false}
	if r.IsActive() {
		t.Error("expected IsActive=false")
	}
}

func TestIsInStock_WhenNilStock(t *testing.T) {
	r := LoyaltyReward{Stock: nil}
	if !r.IsInStock() {
		t.Error("expected IsInStock=true for unlimited stock")
	}
}

func TestIsInStock_WhenPositiveStock(t *testing.T) {
	n := 5
	r := LoyaltyReward{Stock: &n}
	if !r.IsInStock() {
		t.Error("expected IsInStock=true for stock>0")
	}
}

func TestIsInStock_WhenZeroStock(t *testing.T) {
	n := 0
	r := LoyaltyReward{Stock: &n}
	if r.IsInStock() {
		t.Error("expected IsInStock=false for stock=0")
	}
}

func TestValidateAwardInput_Valid(t *testing.T) {
	msg := ValidateAwardInput(AwardInput{OwnerID: 1, CustomerID: 1, Points: 100})
	if msg != "" {
		t.Errorf("expected no error, got %q", msg)
	}
}

func TestValidateAwardInput_ZeroPoints(t *testing.T) {
	msg := ValidateAwardInput(AwardInput{OwnerID: 1, CustomerID: 1, Points: 0})
	if msg == "" {
		t.Error("expected error for zero points")
	}
}

func TestValidateAwardInput_NegativePoints(t *testing.T) {
	msg := ValidateAwardInput(AwardInput{OwnerID: 1, CustomerID: 1, Points: -10})
	if msg == "" {
		t.Error("expected error for negative points")
	}
}

func TestValidateAwardInput_ZeroOwner(t *testing.T) {
	msg := ValidateAwardInput(AwardInput{OwnerID: 0, CustomerID: 1, Points: 10})
	if msg == "" {
		t.Error("expected error for zero owner_id")
	}
}

func TestValidateRedeemInput_Valid(t *testing.T) {
	msg := ValidateRedeemInput(RedeemInput{OwnerID: 1, CustomerID: 1, Points: 50})
	if msg != "" {
		t.Errorf("expected no error, got %q", msg)
	}
}

func TestValidateRedeemInput_ZeroCustomer(t *testing.T) {
	msg := ValidateRedeemInput(RedeemInput{OwnerID: 1, CustomerID: 0, Points: 50})
	if msg == "" {
		t.Error("expected error for zero customer_id")
	}
}
