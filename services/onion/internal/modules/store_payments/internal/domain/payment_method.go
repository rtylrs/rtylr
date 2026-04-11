package domain

import "time"

type PaymentMethod struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"ownerId"`
	Provider  string    `json:"provider"`
	Config    string    `json:"config"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
}

type PaymentMethodInput struct {
	Provider string `json:"provider"`
	Config   string `json:"config"`
	IsActive bool   `json:"isActive"`
}
