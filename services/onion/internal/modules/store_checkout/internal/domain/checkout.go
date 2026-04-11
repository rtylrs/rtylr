package domain

import "time"

type Checkout struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"ownerId"`
	SessionID     string    `json:"sessionId"`
	CustomerEmail string    `json:"customerEmail"`
	Items         string    `json:"items"`
	Subtotal      float64   `json:"subtotal"`
	Tax           float64   `json:"tax"`
	Total         float64   `json:"total"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type CheckoutCreateInput struct {
	SessionID     string  `json:"sessionId"`
	CustomerEmail string  `json:"customerEmail"`
	Items         string  `json:"items"`
	Subtotal      float64 `json:"subtotal"`
	Tax           float64 `json:"tax"`
	Total         float64 `json:"total"`
}

type CheckoutCompleteInput struct {
	ID int64 `json:"id"`
}
