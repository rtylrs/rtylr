package domain

import (
	"errors"
	"time"
)

var (
	ErrExpenseNotFound  = errors.New("expense not found")
	ErrCategoryNotFound = errors.New("expense category not found")
)

type Expense struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	BusinessID  int64     `json:"business_id"`
	CategoryID  *int64    `json:"category_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Description string    `json:"description"`
	Vendor      string    `json:"vendor"`
	ReceiptURL  string    `json:"receipt_url"`
	ExpenseDate time.Time `json:"expense_date"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
	OwnerID     int64
	BusinessID  int64
	CategoryID  *int64
	Amount      float64
	Currency    string
	Description string
	Vendor      string
	ReceiptURL  string
	ExpenseDate time.Time
}

type ExpenseUpdateInput struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	CategoryID  *int64
	Amount      float64
	Currency    string
	Description string
	Vendor      string
	ReceiptURL  string
	ExpenseDate time.Time
}
