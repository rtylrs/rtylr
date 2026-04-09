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
	OwnerID     int64     `json:"ownerId"`
	CategoryID  *int64    `json:"categoryId"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Description string    `json:"description"`
	Vendor      string    `json:"vendor"`
	ReceiptURL  string    `json:"receiptUrl"`
	ExpenseDate time.Time `json:"expenseDate"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ExpenseCategory struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"ownerId"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"createdAt"`
}

type ExpenseInput struct {
	OwnerID     int64
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
	CategoryID  *int64
	Amount      float64
	Currency    string
	Description string
	Vendor      string
	ReceiptURL  string
	ExpenseDate time.Time
}
