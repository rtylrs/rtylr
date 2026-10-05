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
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ExpenseCategory struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	Name       string
	Color      string
	CreatedAt  time.Time
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
