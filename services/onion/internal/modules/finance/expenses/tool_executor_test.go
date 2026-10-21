package expenses

import (
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/finance/expenses/internal/domain"
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

func TestParseListExpensesToolInput(t *testing.T) {
	parsed, err := parseListExpensesToolInput(map[string]any{"status": "approved"})
	if err != nil {
		t.Fatalf("parseListExpensesToolInput returned error: %v", err)
	}
	if parsed.Status != "approved" {
		t.Fatalf("Status = %q, want approved", parsed.Status)
	}
}

func TestParseListExpensesToolInputRejectsBadStatus(t *testing.T) {
	_, err := parseListExpensesToolInput(map[string]any{"status": "paid"})
	if err == nil || !strings.Contains(err.Error(), "status") {
		t.Fatalf("error = %v, want status validation", err)
	}
}

func TestParseGetExpenseToolInputAcceptsAliases(t *testing.T) {
	cases := []map[string]any{
		{"expense_id": int64(8)},
		{"expense": "#8"},
		{"id": "expense-8"},
		{"id": "exp_8"},
	}

	for _, input := range cases {
		parsed, err := parseGetExpenseToolInput(input)
		if err != nil {
			t.Fatalf("parseGetExpenseToolInput(%v) returned error: %v", input, err)
		}
		if parsed.ExpenseID != 8 {
			t.Fatalf("ExpenseID = %d, want 8 for %v", parsed.ExpenseID, input)
		}
	}
}

func TestParseGetExpenseToolInputRejectsMissingExpense(t *testing.T) {
	_, err := parseGetExpenseToolInput(map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "expense_id") {
		t.Fatalf("error = %v, want expense_id validation", err)
	}
}

func TestExpenseToolPayloadFormatsFields(t *testing.T) {
	categoryID := int64(42)
	expenseDate := time.Date(2026, 6, 18, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 6, 18, 10, 30, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := expenseToolPayload(domain.Expense{
		ID:          8,
		CategoryID:  &categoryID,
		Amount:      19.99,
		Currency:    "USD",
		Description: "Office snacks",
		Vendor:      "Corner Shop",
		ExpenseDate: expenseDate,
		Status:      "pending",
		CreatedAt:   createdAt,
	})

	if payload["id"] != int64(8) {
		t.Fatalf("id = %v", payload["id"])
	}
	gotCategoryID, ok := payload["category_id"].(*int64)
	if !ok || gotCategoryID == nil || *gotCategoryID != categoryID {
		t.Fatalf("category_id = %v", payload["category_id"])
	}
	if payload["expense_date"] != "2026-06-18" {
		t.Fatalf("expense_date = %v", payload["expense_date"])
	}
	if payload["created_at"] != "2026-06-18T07:30:00Z" {
		t.Fatalf("created_at = %v", payload["created_at"])
	}
}

func TestExpenseCategoryToolPayloadFormatsFields(t *testing.T) {
	createdAt := time.Date(2026, 6, 18, 9, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := expenseCategoryToolPayload(domain.ExpenseCategory{
		ID:        3,
		Name:      "Supplies",
		Color:     "#ffcc00",
		CreatedAt: createdAt,
	})

	if payload["id"] != int64(3) {
		t.Fatalf("id = %v", payload["id"])
	}
	if payload["name"] != "Supplies" {
		t.Fatalf("name = %v", payload["name"])
	}
	if payload["created_at"] != "2026-06-18T06:00:00Z" {
		t.Fatalf("created_at = %v", payload["created_at"])
	}
}
