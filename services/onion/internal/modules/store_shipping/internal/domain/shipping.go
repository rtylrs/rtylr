package domain

import "time"

type ShippingMethod struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"owner_id"`
	BusinessID    int64     `json:"business_id"`
	Name          string    `json:"name"`
	Price         float64   `json:"price"`
	EstimatedDays int       `json:"estimated_days"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

type ShippingMethodInput struct {
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	EstimatedDays int     `json:"estimated_days"`
	IsActive      bool    `json:"is_active"`
}
