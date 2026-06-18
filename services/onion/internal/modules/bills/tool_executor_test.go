package bills

import (
	"strings"
	"testing"
)

func TestParseRecordBillToolInput(t *testing.T) {
	parsed, err := parseRecordBillToolInput(map[string]any{
		"supplier_id": "42",
		"amount":      "250.75",
		"currency":    "usd",
		"due_date":    "2026-07-15",
		"description": "Monthly stock invoice",
	})
	if err != nil {
		t.Fatalf("parseRecordBillToolInput returned error: %v", err)
	}
	if parsed.SupplierID != 42 {
		t.Fatalf("SupplierID = %d, want 42", parsed.SupplierID)
	}
	if parsed.Amount != 250.75 {
		t.Fatalf("Amount = %v, want 250.75", parsed.Amount)
	}
	if parsed.Currency != "USD" {
		t.Fatalf("Currency = %q, want USD", parsed.Currency)
	}
	if parsed.DueDate != "2026-07-15" {
		t.Fatalf("DueDate = %q, want 2026-07-15", parsed.DueDate)
	}
	if parsed.Status != "pending" {
		t.Fatalf("Status = %q, want pending", parsed.Status)
	}
}

func TestParseRecordBillToolInputAcceptsRFC3339DueDate(t *testing.T) {
	parsed, err := parseRecordBillToolInput(map[string]any{
		"supplier_id": 42,
		"amount":      250.75,
		"due_date":    "2026-07-15T10:30:00Z",
		"status":      "paid",
	})
	if err != nil {
		t.Fatalf("parseRecordBillToolInput returned error: %v", err)
	}
	if parsed.DueDate != "2026-07-15" {
		t.Fatalf("DueDate = %q, want 2026-07-15", parsed.DueDate)
	}
	if parsed.Status != "paid" {
		t.Fatalf("Status = %q, want paid", parsed.Status)
	}
}

func TestParseRecordBillToolInputRejectsMissingSupplier(t *testing.T) {
	_, err := parseRecordBillToolInput(map[string]any{
		"amount":   250.75,
		"due_date": "2026-07-15",
	})
	if err == nil || !strings.Contains(err.Error(), "supplier_id") {
		t.Fatalf("error = %v, want supplier_id validation", err)
	}
}

func TestParseRecordBillToolInputRejectsBadDueDate(t *testing.T) {
	_, err := parseRecordBillToolInput(map[string]any{
		"supplier_id": 42,
		"amount":      250.75,
		"due_date":    "15/07/2026",
	})
	if err == nil || !strings.Contains(err.Error(), "due_date") {
		t.Fatalf("error = %v, want due_date validation", err)
	}
}

func TestParseRecordBillToolInputRejectsBadStatus(t *testing.T) {
	_, err := parseRecordBillToolInput(map[string]any{
		"supplier_id": 42,
		"amount":      250.75,
		"due_date":    "2026-07-15",
		"status":      "draft",
	})
	if err == nil || !strings.Contains(err.Error(), "status") {
		t.Fatalf("error = %v, want status validation", err)
	}
}
