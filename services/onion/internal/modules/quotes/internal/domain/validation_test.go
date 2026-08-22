package domain

import "testing"

func TestValidateQuoteInput_HappyPath(t *testing.T) {
	input := QuoteInput{CustomerName: "Acme", Status: "draft", Subtotal: 100, Total: 100}
	if err := ValidateQuoteInput(input); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateQuoteInput_TotalMustMatchComponents(t *testing.T) {
	input := QuoteInput{CustomerName: "Acme", Subtotal: 100, Discount: 10, Tax: 8.5, Total: 100}
	if err := ValidateQuoteInput(input); err == nil {
		t.Fatal("expected error when total != subtotal - discount + tax")
	}

	consistent := QuoteInput{CustomerName: "Acme", Subtotal: 100, Discount: 10, Tax: 8.5, Total: 98.5}
	if err := ValidateQuoteInput(consistent); err != nil {
		t.Errorf("unexpected error for consistent totals: %v", err)
	}
}

func TestValidateQuoteInput_MissingCustomerName(t *testing.T) {
	input := QuoteInput{Total: 100}
	if err := ValidateQuoteInput(input); err == nil {
		t.Fatal("expected error for missing customer_name")
	}
}

func TestValidateQuoteInput_NegativeTotal(t *testing.T) {
	input := QuoteInput{CustomerName: "Acme", Total: -1}
	if err := ValidateQuoteInput(input); err == nil {
		t.Fatal("expected error for negative total")
	}
}

func TestValidateQuoteInput_NegativeSubtotal(t *testing.T) {
	input := QuoteInput{CustomerName: "Acme", Subtotal: -10}
	if err := ValidateQuoteInput(input); err == nil {
		t.Fatal("expected error for negative subtotal")
	}
}

func TestValidateQuoteInput_NegativeTax(t *testing.T) {
	input := QuoteInput{CustomerName: "Acme", Tax: -5}
	if err := ValidateQuoteInput(input); err == nil {
		t.Fatal("expected error for negative tax")
	}
}

func TestValidateQuoteInput_NegativeDiscount(t *testing.T) {
	input := QuoteInput{CustomerName: "Acme", Discount: -5}
	if err := ValidateQuoteInput(input); err == nil {
		t.Fatal("expected error for negative discount")
	}
}

func TestValidateQuoteInput_InvalidStatus(t *testing.T) {
	input := QuoteInput{CustomerName: "Acme", Status: "bogus"}
	if err := ValidateQuoteInput(input); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestValidateQuoteInput_ValidStatuses(t *testing.T) {
	for status := range ValidQuoteStatuses {
		input := QuoteInput{CustomerName: "Acme", Status: status}
		if err := ValidateQuoteInput(input); err != nil {
			t.Errorf("unexpected error for status=%q: %v", status, err)
		}
	}
}

func TestValidateQuoteInput_BadValidUntilFormat(t *testing.T) {
	input := QuoteInput{CustomerName: "Acme", ValidUntil: "31-12-2026"}
	if err := ValidateQuoteInput(input); err == nil {
		t.Fatal("expected error for bad valid_until format")
	}
}

func TestValidateQuoteInput_ValidUntilAccepted(t *testing.T) {
	input := QuoteInput{CustomerName: "Acme", ValidUntil: "2026-12-31"}
	if err := ValidateQuoteInput(input); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateQuoteInput_EmptyStatusOK(t *testing.T) {
	input := QuoteInput{CustomerName: "Acme"}
	if err := ValidateQuoteInput(input); err != nil {
		t.Errorf("unexpected error for empty status: %v", err)
	}
}
