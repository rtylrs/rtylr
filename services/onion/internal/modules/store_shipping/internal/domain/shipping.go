package domain

import "time"

type ShippingMethod struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"ownerId"`
	Name          string    `json:"name"`
	Price         float64   `json:"price"`
	EstimatedDays int       `json:"estimatedDays"`
	IsActive      bool      `json:"isActive"`
	CreatedAt     time.Time `json:"createdAt"`
}

type ShippingMethodInput struct {
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	EstimatedDays int     `json:"estimatedDays"`
	IsActive      bool    `json:"isActive"`
}
