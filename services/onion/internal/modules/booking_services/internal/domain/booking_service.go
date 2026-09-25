package domain

import (
	"time"

	"github.com/voxire/onion/internal/platform/money"
)

type BookingService struct {
	ID              int64        `json:"id"`
	OwnerID         int64        `json:"owner_id"`
	BusinessID      int64        `json:"business_id"`
	Name            string       `json:"name"`
	Description     string       `json:"description"`
	DurationMinutes int          `json:"duration_minutes"`
	Price           money.Amount `json:"price"`
	Color           string       `json:"color"`
	IsActive        bool         `json:"is_active"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

type BookingServiceInput struct {
	Name            string       `json:"name"`
	Description     string       `json:"description"`
	DurationMinutes int          `json:"duration_minutes"`
	Price           money.Amount `json:"price"`
	Color           string       `json:"color"`
	IsActive        bool         `json:"is_active"`
}
