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
	ID          int64
	OwnerID     int64
	BusinessID  int64
	InvoiceID   int64
	Amount      float64
	Currency    string
	PaymentDate time.Time
	Method      string
	Reference   string
	Notes       string
	CreatedAt   time.Time
}
type PaymentInput struct {
	InvoiceID   int64
	Amount      float64
	Currency    string
	PaymentDate string
	Method      string
	Reference   string
	Notes       string
}
