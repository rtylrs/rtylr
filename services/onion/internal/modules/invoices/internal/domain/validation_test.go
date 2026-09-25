package domain

import (
	"errors"
	"testing"
)

func TestValidateInvoiceInput_HappyPath(t *testing.T) {
	input := InvoiceInput{
		ClientName: "Acme Corp",
		IssueDate:  "2026-01-01",
		DueDate:    "2026-01-31",
		TaxRate:    10,
	}
	if err := ValidateInvoiceInput(input); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateInvoiceInput_MissingClientName(t *testing.T) {
	input := InvoiceInput{
		IssueDate: "2026-01-01",
		DueDate:   "2026-01-31",
	}
	if err := ValidateInvoiceInput(input); err == nil {
		t.Fatal("expected error for missing client_name")
	}
}

func TestValidateInvoiceInput_MissingIssueDate(t *testing.T) {
	input := InvoiceInput{
		ClientName: "Acme",
		DueDate:    "2026-01-31",
	}
	if err := ValidateInvoiceInput(input); err == nil {
		t.Fatal("expected error for missing issue_date")
	}
}

func TestValidateInvoiceInput_MissingDueDate(t *testing.T) {
	input := InvoiceInput{
		ClientName: "Acme",
		IssueDate:  "2026-01-01",
	}
	if err := ValidateInvoiceInput(input); err == nil {
		t.Fatal("expected error for missing due_date")
	}
}

func TestValidateInvoiceInput_DueDateBeforeIssueDate(t *testing.T) {
	input := InvoiceInput{
		ClientName: "Acme",
		IssueDate:  "2026-02-01",
		DueDate:    "2026-01-01",
	}
	err := ValidateInvoiceInput(input)
	if !errors.Is(err, ErrInvalidInvoiceDate) {
		t.Errorf("expected ErrInvalidInvoiceDate, got %v", err)
	}
}

func TestValidateInvoiceInput_SameDayIssueDue(t *testing.T) {
	input := InvoiceInput{
		ClientName: "Acme",
		IssueDate:  "2026-01-15",
		DueDate:    "2026-01-15",
	}
	if err := ValidateInvoiceInput(input); err != nil {
		t.Errorf("same-day issue/due should be valid, got: %v", err)
	}
}

func TestValidateInvoiceInput_BadIssueDateFormat(t *testing.T) {
	input := InvoiceInput{
		ClientName: "Acme",
		IssueDate:  "01-01-2026",
		DueDate:    "2026-01-31",
	}
	if err := ValidateInvoiceInput(input); err == nil {
		t.Fatal("expected error for bad issue_date format")
	}
}

func TestValidateInvoiceInput_NegativeTaxRate(t *testing.T) {
	input := InvoiceInput{
		ClientName: "Acme",
		IssueDate:  "2026-01-01",
		DueDate:    "2026-01-31",
		TaxRate:    -1,
	}
	if err := ValidateInvoiceInput(input); err == nil {
		t.Fatal("expected error for negative tax_rate")
	}
}

func TestValidateInvoiceInput_TaxRateOver100(t *testing.T) {
	input := InvoiceInput{
		ClientName: "Acme",
		IssueDate:  "2026-01-01",
		DueDate:    "2026-01-31",
		TaxRate:    101,
	}
	if err := ValidateInvoiceInput(input); err == nil {
		t.Fatal("expected error for tax_rate > 100")
	}
}

func TestValidateInvoiceInput_NegativeDiscountValue(t *testing.T) {
	input := InvoiceInput{
		ClientName:    "Acme",
		IssueDate:     "2026-01-01",
		DueDate:       "2026-01-31",
		DiscountValue: -5,
	}
	if err := ValidateInvoiceInput(input); err == nil {
		t.Fatal("expected error for negative discount_value")
	}
}

func TestValidateInvoiceInput_BadDiscountType(t *testing.T) {
	input := InvoiceInput{
		ClientName:   "Acme",
		IssueDate:    "2026-01-01",
		DueDate:      "2026-01-31",
		DiscountType: "random",
	}
	if err := ValidateInvoiceInput(input); err == nil {
		t.Fatal("expected error for invalid discount_type")
	}
}

func TestValidateInvoiceInput_ValidDiscountTypeFixed(t *testing.T) {
	input := InvoiceInput{
		ClientName:   "Acme",
		IssueDate:    "2026-01-01",
		DueDate:      "2026-01-31",
		DiscountType: "fixed",
	}
	if err := ValidateInvoiceInput(input); err != nil {
		t.Errorf("unexpected error for discount_type=fixed: %v", err)
	}
}

func TestValidateInvoiceInput_ItemMissingDescription(t *testing.T) {
	input := InvoiceInput{
		ClientName: "Acme",
		IssueDate:  "2026-01-01",
		DueDate:    "2026-01-31",
		Items:      []InvoiceItemInput{{Quantity: 1, UnitPrice: 100}},
	}
	if err := ValidateInvoiceInput(input); err == nil {
		t.Fatal("expected error for item missing description")
	}
}

func TestValidateInvoiceInput_ItemZeroQuantity(t *testing.T) {
	input := InvoiceInput{
		ClientName: "Acme",
		IssueDate:  "2026-01-01",
		DueDate:    "2026-01-31",
		Items:      []InvoiceItemInput{{Description: "Widget", Quantity: 0, UnitPrice: 100}},
	}
	if err := ValidateInvoiceInput(input); err == nil {
		t.Fatal("expected error for zero quantity")
	}
}

func TestValidatePaymentInput_HappyPath(t *testing.T) {
	input := PaymentInput{
		Amount:      500,
		PaymentDate: "2026-01-15",
		Method:      "bank_transfer",
	}
	if err := ValidatePaymentInput(input); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidatePaymentInput_ZeroAmount(t *testing.T) {
	input := PaymentInput{
		Amount:      0,
		PaymentDate: "2026-01-15",
		Method:      "cash",
	}
	if err := ValidatePaymentInput(input); err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestValidatePaymentInput_MissingDate(t *testing.T) {
	input := PaymentInput{
		Amount: 100,
		Method: "cash",
	}
	if err := ValidatePaymentInput(input); err == nil {
		t.Fatal("expected error for missing payment_date")
	}
}

func TestValidatePaymentInput_BadDateFormat(t *testing.T) {
	input := PaymentInput{
		Amount:      100,
		PaymentDate: "15-01-2026",
		Method:      "cash",
	}
	if err := ValidatePaymentInput(input); err == nil {
		t.Fatal("expected error for bad date format")
	}
}

func TestValidatePaymentInput_MissingMethod(t *testing.T) {
	input := PaymentInput{
		Amount:      100,
		PaymentDate: "2026-01-15",
	}
	if err := ValidatePaymentInput(input); err == nil {
		t.Fatal("expected error for missing method")
	}
}
