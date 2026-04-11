package domain

import "time"

type BookingService struct {
	ID              int64     `json:"id"`
	OwnerID         int64     `json:"ownerId"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	DurationMinutes int       `json:"durationMinutes"`
	Price           float64   `json:"price"`
	Color           string    `json:"color"`
	IsActive        bool      `json:"isActive"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type BookingServiceInput struct {
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	DurationMinutes int     `json:"durationMinutes"`
	Price           float64 `json:"price"`
	Color           string  `json:"color"`
	IsActive        bool    `json:"isActive"`
}
