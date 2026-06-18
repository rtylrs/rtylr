package expenses

import (
	"strings"
	"testing"
	"time"
)

func TestParseRecordExpenseToolInput(t *testing.T) {
	parsed, err := parseRecordExpenseToolInput(map[string]any{
		"category_id":  "42",
		"amount":       "19.99",
		"currency":     "usd",
		"description":  "Office snacks",
		"vendor":       "Corner Shop",
		"receipt_url":  "https://example.com/receipt",
		"expense_date": "2026-06-18",
	})
	if err != nil {
		t.Fatalf("parseRecordExpenseToolInput returned error: %v", err)
	}
	if parsed.CategoryID == nil || *parsed.CategoryID != 42 {
		t.Fatalf("CategoryID = %v, want 42", parsed.CategoryID)
	}
	if parsed.Amount != 19.99 {
		t.Fatalf("Amount = %v, want 19.99", parsed.Amount)
	}
	if parsed.Currency != "USD" {
		t.Fatalf("Currency = %q, want USD", parsed.Currency)
	}
	if parsed.Description != "Office snacks" {
		t.Fatalf("Description = %q, want Office snacks", parsed.Description)
	}
	if parsed.Vendor != "Corner Shop" {
		t.Fatalf("Vendor = %q, want Corner Shop", parsed.Vendor)
	}
	if got := parsed.ExpenseDate.Format("2006-01-02"); got != "2026-06-18" {
		t.Fatalf("ExpenseDate = %q, want 2026-06-18", got)
	}
}

func TestParseRecordExpenseToolInputRejectsFutureDate(t *testing.T) {
	_, err := parseRecordExpenseToolInput(map[string]any{
		"amount":       10,
		"description":  "Future expense",
		"expense_date": time.Now().AddDate(0, 0, 3).Format("2006-01-02"),
	})
	if err == nil || !strings.Contains(err.Error(), "must not be in the future") {
		t.Fatalf("error = %v, want future date validation", err)
	}
}
