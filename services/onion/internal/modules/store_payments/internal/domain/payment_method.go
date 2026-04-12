package domain

import "time"

type PaymentMethod struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"owner_id"`
	Provider  string    `json:"provider"`
	Config    string    `json:"config"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type PaymentMethodInput struct {
	Provider string `json:"provider"`
	Config   string `json:"config"`
	IsActive bool   `json:"is_active"`
}
