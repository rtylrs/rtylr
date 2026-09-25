package domain

import (
	"errors"
	"time"
)

var (
	ErrQuoteNotFound         = errors.New("quote not found")
	ErrQuoteNotEditable      = errors.New("quote is not editable")
	ErrQuoteAlreadyConverted = errors.New("quote already converted")
	ErrQuoteTotalMismatch    = errors.New("quote total does not equal subtotal - discount + tax")
)

type Quote struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"owner_id"`
	BusinessID    int64     `json:"business_id"`
	CustomerName  string    `json:"customer_name"`
	CustomerEmail string    `json:"customer_email"`
	Status        string    `json:"status"`
	Items         string    `json:"items"`
	Subtotal      float64   `json:"subtotal"`
	Tax           float64   `json:"tax"`
	Discount      float64   `json:"discount"`
	Total         float64   `json:"total"`
	ValidUntil    time.Time `json:"valid_until"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
type QuoteInput struct {
	CustomerName  string  `json:"customer_name"`
	CustomerEmail string  `json:"customer_email"`
	Status        string  `json:"status"`
	Items         string  `json:"items"`
	Subtotal      float64 `json:"subtotal"`
	Tax           float64 `json:"tax"`
	Discount      float64 `json:"discount"`
	Total         float64 `json:"total"`
	ValidUntil    string  `json:"valid_until"`
	Notes         string  `json:"notes"`
}

// IsExpired reports whether the quote has passed its ValidUntil date and has
// not been accepted or declined.
func (q *Quote) IsExpired() bool {
	if q.Status == "accepted" || q.Status == "declined" {
		return false
	}
	if q.ValidUntil.IsZero() {
		return false
	}
	return time.Now().After(q.ValidUntil)
}

// IsEditable reports whether the quote can still be modified. Only draft and
// sent quotes may be edited.
func (q *Quote) IsEditable() bool {
	return q.Status == "draft" || q.Status == "sent"
}

// NetTotal returns the quote total after applying the discount and adding tax.
// This mirrors the stored Total but can be used for verification.
func (q *Quote) NetTotal() float64 {
	return q.Subtotal - q.Discount + q.Tax
}
