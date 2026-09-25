package invoices

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
)

func TestParseListInvoicesToolInput(t *testing.T) {
	parsed, err := parseListInvoicesToolInput(map[string]any{
		"status":        "overdue",
		"customer_name": "Maya Haddad",
		"date_from":     "2026-06-01T10:30:00Z",
		"date_to":       "2026-06-30",
		"page":          "2",
		"limit":         100,
	})
	if err != nil {
		t.Fatalf("parseListInvoicesToolInput returned error: %v", err)
	}
	if parsed.Status != domain.StatusOverdue {
		t.Fatalf("Status = %q, want overdue", parsed.Status)
	}
	if parsed.Search != "Maya Haddad" {
		t.Fatalf("Search = %q, want Maya Haddad", parsed.Search)
	}
	if parsed.DateFrom != "2026-06-01" {
		t.Fatalf("DateFrom = %q, want 2026-06-01", parsed.DateFrom)
	}
	if parsed.DateTo != "2026-06-30" {
		t.Fatalf("DateTo = %q, want 2026-06-30", parsed.DateTo)
	}
	if parsed.Page != 2 {
		t.Fatalf("Page = %d, want 2", parsed.Page)
	}
	if parsed.PageSize != maxListInvoicesToolPageSize {
		t.Fatalf("PageSize = %d, want cap %d", parsed.PageSize, maxListInvoicesToolPageSize)
	}
}

func TestParseListInvoicesToolInputUsesDefaults(t *testing.T) {
	parsed, err := parseListInvoicesToolInput(map[string]any{})
	if err != nil {
		t.Fatalf("parseListInvoicesToolInput returned error: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d, want 1", parsed.Page)
	}
	if parsed.PageSize != defaultListInvoicesToolPageSize {
		t.Fatalf("PageSize = %d, want %d", parsed.PageSize, defaultListInvoicesToolPageSize)
	}
}

func TestParseListInvoicesToolInputRejectsInvalidStatus(t *testing.T) {
	_, err := parseListInvoicesToolInput(map[string]any{"status": "stuck"})
	if err == nil || !strings.Contains(err.Error(), "status") {
		t.Fatalf("error = %v, want status validation", err)
	}
}

func TestParseListInvoicesToolInputRejectsInvalidDateRange(t *testing.T) {
	_, err := parseListInvoicesToolInput(map[string]any{
		"date_from": "2026-06-30",
		"date_to":   "2026-06-01",
	})
	if err == nil || !strings.Contains(err.Error(), "date_to") {
		t.Fatalf("error = %v, want date range validation", err)
	}
}

func TestParseGetInvoiceToolInputAcceptsAliases(t *testing.T) {
	for _, input := range []map[string]any{
		{"invoice_id": "42"},
		{"invoice": 42},
		{"id": float64(42)},
	} {
		invoiceID, err := parseGetInvoiceToolInput(input)
		if err != nil {
			t.Fatalf("parseGetInvoiceToolInput(%v) returned error: %v", input, err)
		}
		if invoiceID != 42 {
			t.Fatalf("invoiceID = %d, want 42", invoiceID)
		}
	}
}

func TestParseGetInvoiceToolInputRejectsMissingInvoice(t *testing.T) {
	_, err := parseGetInvoiceToolInput(map[string]any{"client_name": "Maya Haddad"})
	if err == nil || !strings.Contains(err.Error(), "invoice_id") {
		t.Fatalf("error = %v, want invoice_id validation", err)
	}
}

func TestInvoiceToolPayload(t *testing.T) {
	paidDate := time.Date(2026, 6, 19, 9, 0, 0, 0, time.UTC)
	payload := invoiceToolPayload(domain.Invoice{
		ID:            42,
		InvoiceNumber: "INV-2026-0042",
		ClientName:    "Maya Haddad",
		Status:        domain.StatusPaid,
		IssueDate:     time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		DueDate:       time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
		PaidDate:      &paidDate,
		Currency:      "USD",
		Total:         150,
		AmountPaid:    150,
		AmountDue:     0,
		CreatedAt:     time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
	})
	if payload["invoice_number"] != "INV-2026-0042" {
		t.Fatalf("invoice_number = %v, want INV-2026-0042", payload["invoice_number"])
	}
	if payload["paid_date"] != "2026-06-19" {
		t.Fatalf("paid_date = %v, want 2026-06-19", payload["paid_date"])
	}
	if payload["created_at"] != "2026-06-01T10:00:00Z" {
		t.Fatalf("created_at = %v, want formatted timestamp", payload["created_at"])
	}
}

func TestInvoiceDetailToolPayloadIncludesItemsAndPayments(t *testing.T) {
	paidDate := time.Date(2026, 6, 19, 9, 0, 0, 0, time.UTC)
	payload := invoiceDetailToolPayload(domain.Invoice{
		ID:             42,
		InvoiceNumber:  "INV-2026-0042",
		ClientName:     "Maya Haddad",
		ClientEmail:    "maya@example.com",
		Status:         domain.StatusPartial,
		IssueDate:      time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		DueDate:        time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
		PaidDate:       &paidDate,
		Currency:       "USD",
		Subtotal:       200,
		DiscountAmount: 10,
		TaxAmount:      5,
		Total:          195,
		AmountPaid:     100,
		AmountDue:      95,
		Items: []domain.InvoiceItem{
			{ID: 7, InvoiceID: 42, Description: "Monthly support", Quantity: 2, Unit: "hour", UnitPrice: 100, DiscountPct: 5, Total: 190},
		},
		Payments: []domain.Payment{
			{ID: 9, InvoiceID: 42, Amount: 100, Currency: "USD", PaymentDate: paidDate, Method: domain.PaymentMethodCard, Reference: "TX-123", CreatedAt: paidDate},
		},
		CreatedAt: time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 6, 2, 11, 0, 0, 0, time.UTC),
	})

	if payload["client_email"] != "maya@example.com" {
		t.Fatalf("client_email = %v", payload["client_email"])
	}
	if payload["amount_due"] != 95.0 {
		t.Fatalf("amount_due = %v", payload["amount_due"])
	}
	items, ok := payload["items"].([]map[string]any)
	if !ok {
		t.Fatalf("items type = %T", payload["items"])
	}
	if len(items) != 1 || items[0]["description"] != "Monthly support" {
		t.Fatalf("items = %#v", items)
	}
	payments, ok := payload["payments"].([]map[string]any)
	if !ok {
		t.Fatalf("payments type = %T", payload["payments"])
	}
	if len(payments) != 1 || payments[0]["reference"] != "TX-123" {
		t.Fatalf("payments = %#v", payments)
	}
}

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

func TestParseRecordPaymentToolInput(t *testing.T) {
	parsed, err := parseRecordPaymentToolInput(map[string]any{
		"invoice_id":     "42",
		"payment_amount": "125.50",
		"currency":       "usd",
		"paid_at":        "2026-07-15T10:30:00Z",
		"method":         "bank transfer",
		"transaction_id": "TX-123",
		"notes":          "First installment",
	}, time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("parseRecordPaymentToolInput returned error: %v", err)
	}
	if parsed.InvoiceID != 42 {
		t.Fatalf("InvoiceID = %d, want 42", parsed.InvoiceID)
	}
	if parsed.Amount != 125.50 {
		t.Fatalf("Amount = %v, want 125.50", parsed.Amount)
	}
	if parsed.Currency != "USD" {
		t.Fatalf("Currency = %q, want USD", parsed.Currency)
	}
	if parsed.PaymentDate != "2026-07-15" {
		t.Fatalf("PaymentDate = %q, want 2026-07-15", parsed.PaymentDate)
	}
	if parsed.Method != domain.PaymentMethodBankTransfer {
		t.Fatalf("Method = %q, want bank_transfer", parsed.Method)
	}
	if parsed.Reference != "TX-123" {
		t.Fatalf("Reference = %q, want TX-123", parsed.Reference)
	}
}

func TestParseRecordPaymentToolInputDefaultsDateAndMethod(t *testing.T) {
	parsed, err := parseRecordPaymentToolInput(map[string]any{
		"id":     42,
		"amount": 125.50,
	}, time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("parseRecordPaymentToolInput returned error: %v", err)
	}
	if parsed.PaymentDate != "2026-06-19" {
		t.Fatalf("PaymentDate = %q, want 2026-06-19", parsed.PaymentDate)
	}
	if parsed.Method != domain.PaymentMethodCash {
		t.Fatalf("Method = %q, want cash", parsed.Method)
	}
}

func TestParseRecordPaymentToolInputRejectsMissingInvoice(t *testing.T) {
	_, err := parseRecordPaymentToolInput(map[string]any{
		"amount": 125.50,
	}, time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
	if err == nil || !strings.Contains(err.Error(), "invoice_id") {
		t.Fatalf("error = %v, want invoice_id validation", err)
	}
}

func TestParseRecordPaymentToolInputRejectsInvalidAmount(t *testing.T) {
	_, err := parseRecordPaymentToolInput(map[string]any{
		"invoice_id": 42,
		"amount":     0,
	}, time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
	if !errors.Is(err, domain.ErrInvalidPaymentAmount) {
		t.Fatalf("error = %v, want ErrInvalidPaymentAmount", err)
	}
}

func TestParseRecordPaymentToolInputRejectsBadMethod(t *testing.T) {
	_, err := parseRecordPaymentToolInput(map[string]any{
		"invoice_id": 42,
		"amount":     125.50,
		"method":     "crypto",
	}, time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
	if err == nil || !strings.Contains(err.Error(), "method") {
		t.Fatalf("error = %v, want method validation", err)
	}
}

func TestParseRecordPaymentToolInputRejectsBadCurrency(t *testing.T) {
	_, err := parseRecordPaymentToolInput(map[string]any{
		"invoice_id": 42,
		"amount":     125.50,
		"currency":   "USDD",
	}, time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
	if err == nil || !strings.Contains(err.Error(), "currency") {
		t.Fatalf("error = %v, want currency validation", err)
	}
}

func TestParseSendInvoiceToolInput(t *testing.T) {
	invoiceID, err := parseSendInvoiceToolInput(map[string]any{
		"invoice": "42",
	})
	if err != nil {
		t.Fatalf("parseSendInvoiceToolInput returned error: %v", err)
	}
	if invoiceID != 42 {
		t.Fatalf("invoiceID = %d, want 42", invoiceID)
	}
}

func TestParseSendInvoiceToolInputRejectsMissingInvoice(t *testing.T) {
	_, err := parseSendInvoiceToolInput(map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "invoice_id") {
		t.Fatalf("error = %v, want invoice_id validation", err)
	}
}

func TestParseSendInvoiceToolInputRejectsBadInvoice(t *testing.T) {
	_, err := parseSendInvoiceToolInput(map[string]any{
		"invoice_id": "abc",
	})
	if err == nil || !strings.Contains(err.Error(), "invoice_id") {
		t.Fatalf("error = %v, want invoice_id validation", err)
	}
}
