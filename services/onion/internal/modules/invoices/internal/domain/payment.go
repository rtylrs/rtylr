package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidPaymentAmount  = errors.New("payment amount must be greater than zero")
	ErrPaymentExceedsBalance = errors.New("payment cannot exceed the remaining amount due")
)

type Payment struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	BusinessID  int64     `json:"business_id"`
	InvoiceID   int64     `json:"invoice_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	PaymentDate time.Time `json:"payment_date"`
	Method      string    `json:"method"`
	Reference   string    `json:"reference"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
}
type PaymentInput struct {
	InvoiceID   int64   `json:"invoice_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	PaymentDate string  `json:"payment_date"`
	Method      string  `json:"method"`
	Reference   string  `json:"reference"`
	Notes       string  `json:"notes"`
}
