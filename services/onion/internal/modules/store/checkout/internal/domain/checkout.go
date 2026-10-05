package domain

import "time"

type Checkout struct {
	ID            int64
	OwnerID       int64
	BusinessID    int64
	SessionID     string
	CustomerEmail string
	Items         string
	Subtotal      float64
	Tax           float64
	Total         float64
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CheckoutCreateInput struct {
	SessionID     string
	CustomerEmail string
	Items         string
	Subtotal      float64
	Tax           float64
	Total         float64
}

// CentsToFloat converts integer cents to standard float amount.
func CentsToFloat(cents int64) float64 {
	return float64(cents) / 100.0
}
