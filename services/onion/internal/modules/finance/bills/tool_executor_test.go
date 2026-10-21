package bills

import (
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/finance/bills/internal/domain"
	"github.com/voxire/onion/internal/platform/money"
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
	if parsed.Amount != money.Amount(25075) {
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

func TestParseListBillsToolInput(t *testing.T) {
	parsed, err := parseListBillsToolInput(map[string]any{"status": "paid"})
	if err != nil {
		t.Fatalf("parseListBillsToolInput returned error: %v", err)
	}
	if parsed.Status != "paid" {
		t.Fatalf("Status = %q, want paid", parsed.Status)
	}
}

func TestParseListBillsToolInputRejectsBadStatus(t *testing.T) {
	_, err := parseListBillsToolInput(map[string]any{"status": "draft"})
	if err == nil || !strings.Contains(err.Error(), "status") {
		t.Fatalf("error = %v, want status validation", err)
	}
}

func TestBillToolPayloadFormatsFields(t *testing.T) {
	dueDate := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 6, 18, 14, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := billToolPayload(domain.Bill{
		ID:          5,
		SupplierID:  42,
		Amount:      money.Amount(25075),
		Currency:    "USD",
		DueDate:     dueDate,
		Status:      "pending",
		Description: "Monthly stock invoice",
		CreatedAt:   createdAt,
	})

	if payload["id"] != int64(5) {
		t.Fatalf("id = %v", payload["id"])
	}
	if payload["supplier_id"] != int64(42) {
		t.Fatalf("supplier_id = %v", payload["supplier_id"])
	}
	if payload["due_date"] != "2026-07-15" {
		t.Fatalf("due_date = %v", payload["due_date"])
	}
	if payload["created_at"] != "2026-06-18T11:00:00Z" {
		t.Fatalf("created_at = %v", payload["created_at"])
	}
}
