package quotes

import (
	"strings"
	"testing"
)

func TestParseCreateQuoteToolInput(t *testing.T) {
	parsed, err := parseCreateQuoteToolInput(map[string]any{
		"customer_name":  "Maya Haddad",
		"customer_email": "maya@example.com",
		"items": []any{
			map[string]any{"description": "Protein bundle", "quantity": 2, "unit_price": 50},
		},
		"subtotal":    "100",
		"tax":         11,
		"discount":    5,
		"valid_until": "2026-07-15",
		"notes":       "Includes delivery",
	})
	if err != nil {
		t.Fatalf("parseCreateQuoteToolInput returned error: %v", err)
	}
	if parsed.CustomerName != "Maya Haddad" {
		t.Fatalf("CustomerName = %q, want Maya Haddad", parsed.CustomerName)
	}
	if parsed.Status != "draft" {
		t.Fatalf("Status = %q, want draft", parsed.Status)
	}
	if parsed.Total != 106 {
		t.Fatalf("Total = %v, want 106", parsed.Total)
	}
	if parsed.ValidUntil != "2026-07-15" {
		t.Fatalf("ValidUntil = %q, want 2026-07-15", parsed.ValidUntil)
	}
	if !strings.Contains(parsed.Items, "Protein bundle") {
		t.Fatalf("Items = %q, want serialized item", parsed.Items)
	}
}

func TestParseCreateQuoteToolInputKeepsProvidedTotal(t *testing.T) {
	parsed, err := parseCreateQuoteToolInput(map[string]any{
		"client_name": "Maya Haddad",
		"items":       "Monthly subscription package",
		"subtotal":    100,
		"tax":         11,
		"discount":    5,
		"total":       106,
		"status":      "sent",
	})
	if err != nil {
		t.Fatalf("parseCreateQuoteToolInput returned error: %v", err)
	}
	if parsed.Total != 106 {
		t.Fatalf("Total = %v, want provided total 106", parsed.Total)
	}
	if parsed.Status != "sent" {
		t.Fatalf("Status = %q, want sent", parsed.Status)
	}
	if !strings.Contains(parsed.Items, "Monthly subscription package") {
		t.Fatalf("Items = %q, want text item serialized", parsed.Items)
	}
}

func TestParseCreateQuoteToolInputAcceptsRFC3339ValidUntil(t *testing.T) {
	parsed, err := parseCreateQuoteToolInput(map[string]any{
		"customer_name": "Maya Haddad",
		"valid_until":   "2026-07-15T10:30:00Z",
	})
	if err != nil {
		t.Fatalf("parseCreateQuoteToolInput returned error: %v", err)
	}
	if parsed.ValidUntil != "2026-07-15" {
		t.Fatalf("ValidUntil = %q, want 2026-07-15", parsed.ValidUntil)
	}
}

func TestParseCreateQuoteToolInputRejectsMissingCustomer(t *testing.T) {
	_, err := parseCreateQuoteToolInput(map[string]any{"subtotal": 100})
	if err == nil || !strings.Contains(err.Error(), "customer_name") {
		t.Fatalf("error = %v, want customer validation", err)
	}
}

func TestParseCreateQuoteToolInputRejectsBadStatus(t *testing.T) {
	_, err := parseCreateQuoteToolInput(map[string]any{
		"customer_name": "Maya Haddad",
		"status":        "ready",
	})
	if err == nil || !strings.Contains(err.Error(), "status") {
		t.Fatalf("error = %v, want status validation", err)
	}
}

func TestParseCreateQuoteToolInputRejectsNegativeDiscount(t *testing.T) {
	_, err := parseCreateQuoteToolInput(map[string]any{
		"customer_name": "Maya Haddad",
		"discount":      -1,
	})
	if err == nil || !strings.Contains(err.Error(), "discount") {
		t.Fatalf("error = %v, want discount validation", err)
	}
}

func TestParseCreateQuoteToolInputRejectsInconsistentTotal(t *testing.T) {
	_, err := parseCreateQuoteToolInput(map[string]any{
		"client_name": "Maya Haddad",
		"items":       "Monthly subscription package",
		"subtotal":    100,
		"tax":         11,
		"discount":    5,
		"total":       120,
		"status":      "sent",
	})
	if err == nil {
		t.Fatal("expected error when provided total does not match subtotal - discount + tax")
	}
}
