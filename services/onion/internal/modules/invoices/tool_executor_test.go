package invoices

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
)

func TestParseCreateInvoiceToolInput(t *testing.T) {
	parsed, err := parseCreateInvoiceToolInput(map[string]any{
		"customer_name":  "Maya Haddad",
		"customer_email": "maya@example.com",
		"issue_date":     "2026-07-15T10:30:00Z",
		"due_date":       "2026-07-30",
		"currency":       "usd",
		"discount_type":  "percent",
		"discount":       "5",
		"tax_rate":       11,
		"line_items": []any{
			map[string]any{"name": "Monthly support", "qty": 2, "price": "50", "unit": "hour"},
		},
	}, time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("parseCreateInvoiceToolInput returned error: %v", err)
	}
	if parsed.ClientName != "Maya Haddad" {
		t.Fatalf("ClientName = %q, want Maya Haddad", parsed.ClientName)
	}
	if parsed.ClientEmail != "maya@example.com" {
		t.Fatalf("ClientEmail = %q, want maya@example.com", parsed.ClientEmail)
	}
	if parsed.IssueDate != "2026-07-15" {
		t.Fatalf("IssueDate = %q, want 2026-07-15", parsed.IssueDate)
	}
	if parsed.DueDate != "2026-07-30" {
		t.Fatalf("DueDate = %q, want 2026-07-30", parsed.DueDate)
	}
	if parsed.Currency != "USD" {
		t.Fatalf("Currency = %q, want USD", parsed.Currency)
	}
	if parsed.DiscountType != "percentage" {
		t.Fatalf("DiscountType = %q, want percentage", parsed.DiscountType)
	}
	if parsed.DiscountValue != 5 {
		t.Fatalf("DiscountValue = %v, want 5", parsed.DiscountValue)
	}
	if len(parsed.Items) != 1 {
		t.Fatalf("Items length = %d, want 1", len(parsed.Items))
	}
	item := parsed.Items[0]
	if item.Description != "Monthly support" {
		t.Fatalf("Description = %q, want Monthly support", item.Description)
	}
	if item.Quantity != 2 {
		t.Fatalf("Quantity = %v, want 2", item.Quantity)
	}
	if item.UnitPrice != 50 {
		t.Fatalf("UnitPrice = %v, want 50", item.UnitPrice)
	}
}

func TestParseCreateInvoiceToolInputUsesDefaultsAndSingleItem(t *testing.T) {
	parsed, err := parseCreateInvoiceToolInput(map[string]any{
		"client_name": "Maya Haddad",
		"description": "Setup package",
		"amount":      "250",
		"due_in_days": 14,
	}, time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("parseCreateInvoiceToolInput returned error: %v", err)
	}
	if parsed.IssueDate != "2026-06-19" {
		t.Fatalf("IssueDate = %q, want 2026-06-19", parsed.IssueDate)
	}
	if parsed.DueDate != "2026-07-03" {
		t.Fatalf("DueDate = %q, want 2026-07-03", parsed.DueDate)
	}
	if parsed.DiscountType != "fixed" {
		t.Fatalf("DiscountType = %q, want fixed", parsed.DiscountType)
	}
	if len(parsed.Items) != 1 {
		t.Fatalf("Items length = %d, want 1", len(parsed.Items))
	}
	if parsed.Items[0].Description != "Setup package" {
		t.Fatalf("Description = %q, want Setup package", parsed.Items[0].Description)
	}
	if parsed.Items[0].Quantity != 1 {
		t.Fatalf("Quantity = %v, want 1", parsed.Items[0].Quantity)
	}
	if parsed.Items[0].UnitPrice != 250 {
		t.Fatalf("UnitPrice = %v, want 250", parsed.Items[0].UnitPrice)
	}
}

func TestParseCreateInvoiceToolInputAcceptsJSONItems(t *testing.T) {
	parsed, err := parseCreateInvoiceToolInput(map[string]any{
		"client_name": "Maya Haddad",
		"issue_date":  "2026-06-19",
		"items":       `[{"description":"Protein bundle","quantity":3,"unit_price":25,"discount_pct":10}]`,
	}, time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("parseCreateInvoiceToolInput returned error: %v", err)
	}
	if parsed.DueDate != "2026-06-26" {
		t.Fatalf("DueDate = %q, want default 2026-06-26", parsed.DueDate)
	}
	if len(parsed.Items) != 1 {
		t.Fatalf("Items length = %d, want 1", len(parsed.Items))
	}
	if parsed.Items[0].DiscountPct != 10 {
		t.Fatalf("DiscountPct = %v, want 10", parsed.Items[0].DiscountPct)
	}
}

func TestParseCreateInvoiceToolInputRejectsMissingClient(t *testing.T) {
	_, err := parseCreateInvoiceToolInput(map[string]any{
		"description": "Setup package",
		"amount":      250,
	}, time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
	if err == nil || !strings.Contains(err.Error(), "client_name") {
		t.Fatalf("error = %v, want client_name validation", err)
	}
}

func TestParseCreateInvoiceToolInputRejectsMissingItems(t *testing.T) {
	_, err := parseCreateInvoiceToolInput(map[string]any{
		"client_name": "Maya Haddad",
	}, time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
	if err == nil || !strings.Contains(err.Error(), "item") {
		t.Fatalf("error = %v, want item validation", err)
	}
}

func TestParseCreateInvoiceToolInputRejectsDueDateBeforeIssueDate(t *testing.T) {
	_, err := parseCreateInvoiceToolInput(map[string]any{
		"client_name": "Maya Haddad",
		"issue_date":  "2026-06-20",
		"due_date":    "2026-06-19",
		"description": "Setup package",
		"amount":      250,
	}, time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
	if !errors.Is(err, domain.ErrInvalidInvoiceDate) {
		t.Fatalf("error = %v, want ErrInvalidInvoiceDate", err)
	}
}

func TestParseCreateInvoiceToolInputRejectsBadItemDiscount(t *testing.T) {
	_, err := parseCreateInvoiceToolInput(map[string]any{
		"client_name": "Maya Haddad",
		"items": []any{
			map[string]any{"description": "Setup package", "unit_price": 250, "discount_pct": 101},
		},
	}, time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
	if err == nil || !strings.Contains(err.Error(), "discount_pct") {
		t.Fatalf("error = %v, want discount validation", err)
	}
}
