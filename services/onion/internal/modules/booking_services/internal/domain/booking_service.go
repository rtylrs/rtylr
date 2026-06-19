package domain

import "time"

type BookingService struct {
	ID              int64     `json:"id"`
	OwnerID         int64     `json:"owner_id"`
	BusinessID      int64     `json:"business_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	DurationMinutes int       `json:"duration_minutes"`
	Price           float64   `json:"price"`
	Color           string    `json:"color"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type BookingServiceInput struct {
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	DurationMinutes int     `json:"duration_minutes"`
	Price           float64 `json:"price"`
	Color           string  `json:"color"`
	IsActive        bool    `json:"is_active"`
}
