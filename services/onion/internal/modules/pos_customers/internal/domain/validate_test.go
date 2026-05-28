package domain

import (
	"strings"
	"testing"
)

func TestValidateSearchQuery_AcceptsEmpty(t *testing.T) {
	if err := ValidateSearchQuery(""); err != nil {
		t.Fatalf("unexpected error for empty query: %v", err)
	}
}

func TestValidateSearchQuery_AcceptsNormalQuery(t *testing.T) {
	if err := ValidateSearchQuery("Jamie"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateSearchQuery_RejectsOverlyLongQuery(t *testing.T) {
	long := strings.Repeat("x", 201)
	if err := ValidateSearchQuery(long); err == nil {
		t.Fatal("expected error for 201-char query")
	}
}

func TestValidateSearchQuery_AcceptsMaxLength(t *testing.T) {
	exact := strings.Repeat("a", 200)
	if err := ValidateSearchQuery(exact); err != nil {
		t.Fatalf("unexpected error for 200-char query: %v", err)
	}
}

func TestCustomer_DisplayName_FirstAndLast(t *testing.T) {
	c := Customer{FirstName: "Jamie", LastName: "Smith"}
	if got := c.DisplayName(); got != "Jamie Smith" {
		t.Errorf("DisplayName = %q, want 'Jamie Smith'", got)
	}
}

func TestCustomer_DisplayName_FirstOnly(t *testing.T) {
	c := Customer{FirstName: "Alex", LastName: ""}
	if got := c.DisplayName(); got != "Alex" {
		t.Errorf("DisplayName = %q, want 'Alex'", got)
	}
}

func TestCustomer_HasContactInfo_Email(t *testing.T) {
	c := Customer{Email: "test@example.com"}
	if !c.HasContactInfo() {
		t.Error("expected HasContactInfo to be true when Email is set")
	}
}

func TestCustomer_HasContactInfo_Phone(t *testing.T) {
	c := Customer{Phone: "555-1234"}
	if !c.HasContactInfo() {
		t.Error("expected HasContactInfo to be true when Phone is set")
	}
}

func TestCustomer_HasContactInfo_None(t *testing.T) {
	c := Customer{FirstName: "No Contact"}
	if c.HasContactInfo() {
		t.Error("expected HasContactInfo to be false when no contact info is set")
	}
}
