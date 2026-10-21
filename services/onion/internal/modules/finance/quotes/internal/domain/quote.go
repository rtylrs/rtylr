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
	ID            int64
	OwnerID       int64
	BusinessID    int64
	CustomerName  string
	CustomerEmail string
	Status        string
	Items         string
	Subtotal      float64
	Tax           float64
	Discount      float64
	Total         float64
	ValidUntil    time.Time
	Notes         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
type QuoteInput struct {
	CustomerName  string
	CustomerEmail string
	Status        string
	Items         string
	Subtotal      float64
	Tax           float64
	Discount      float64
	Total         float64
	ValidUntil    string
	Notes         string
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
