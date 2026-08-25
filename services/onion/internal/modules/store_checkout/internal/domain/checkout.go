package domain

import "time"

type Checkout struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"owner_id"`
	BusinessID    int64     `json:"business_id"`
	SessionID     string    `json:"session_id"`
	CustomerEmail string    `json:"customer_email"`
	Items         string    `json:"items"`
	Subtotal      float64   `json:"subtotal"`
	Tax           float64   `json:"tax"`
	Total         float64   `json:"total"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CheckoutCreateInput struct {
	SessionID     string  `json:"session_id"`
	CustomerEmail string  `json:"customer_email"`
	Items         string  `json:"items"`
	Subtotal      float64 `json:"subtotal"`
	Tax           float64 `json:"tax"`
	Total         float64 `json:"total"`
}

type CheckoutCompleteInput struct {
	ID int64 `json:"id"`
}

// CentsToFloat converts integer cents to standard float amount.
func CentsToFloat(cents int64) float64 {
	return float64(cents) / 100.0
}
