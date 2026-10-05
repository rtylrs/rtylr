package domain

import (
	"errors"
	"time"
)

var ErrExpenseNotFound = errors.New("expense not found")

type Expense struct {
	ID              int64
	OwnerID         int64
	BusinessID      int64
	CategoryID      int64
	Description     string
	Amount          float64
	Currency        string
	ExpenseDate     time.Time
	Vendor          string
	PaymentMethod   string
	ReceiptURL      string
	IsTaxDeductible bool
	TaxAmount       float64
	Notes           string
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ExpenseInput struct {
	CategoryID      int64
	Description     string
	Amount          float64
	Currency        string
	ExpenseDate     string
	Vendor          string
	PaymentMethod   string
	ReceiptURL      string
	IsTaxDeductible bool
	TaxAmount       float64
	Notes           string
}
