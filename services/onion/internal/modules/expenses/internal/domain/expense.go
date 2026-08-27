package domain

import (
	"errors"
	"time"
)

var (
	ErrExpenseNotFound          = errors.New("expense not found")
	ErrCategoryNotFound         = errors.New("expense category not found")
	ErrInvalidStatusTransition  = errors.New("invalid expense status transition")
	ErrApprovedExpenseImmutable = errors.New("approved expense is immutable")
)

// ValidateExpenseStatusTransition allows pending expenses to be approved or
// rejected. Approved and rejected are terminal review outcomes; they may be
// repeated idempotently but cannot be changed into another outcome.
func ValidateExpenseStatusTransition(current, next string) error {
	if current == next {
		return nil
	}
	if current == "pending" && (next == "approved" || next == "rejected") {
		return nil
	}
	return ErrInvalidStatusTransition
}

type Expense struct {
	ID              int64     `json:"id"`
	OwnerID         int64     `json:"owner_id"`
	BusinessID      int64     `json:"business_id"`
	CategoryID      *int64    `json:"category_id"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	Description     string    `json:"description"`
	Vendor          string    `json:"vendor"`
	PaymentMethod   string    `json:"payment_method"`
	ReceiptURL      string    `json:"receipt_url"`
	IsTaxDeductible bool      `json:"is_tax_deductible"`
	TaxAmount       float64   `json:"tax_amount"`
	Notes           string    `json:"notes"`
	ExpenseDate     time.Time `json:"expense_date"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ExpenseCategory struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	Name       string    `json:"name"`
	Color      string    `json:"color"`
	CreatedAt  time.Time `json:"created_at"`
}

type ExpenseInput struct {
	OwnerID         int64
	BusinessID      int64
	CategoryID      *int64
	Amount          float64
	Currency        string
	Description     string
	Vendor          string
	PaymentMethod   string
	ReceiptURL      string
	IsTaxDeductible bool
	TaxAmount       float64
	Notes           string
	ExpenseDate     time.Time
}

type ExpenseUpdateInput struct {
	ID              int64
	OwnerID         int64
	BusinessID      int64
	CategoryID      *int64
	Amount          float64
	Currency        string
	Description     string
	Vendor          string
	PaymentMethod   string
	ReceiptURL      string
	IsTaxDeductible bool
	TaxAmount       float64
	Notes           string
	ExpenseDate     time.Time
}
