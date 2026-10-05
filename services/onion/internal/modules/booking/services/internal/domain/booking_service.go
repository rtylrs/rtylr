package domain

import (
	"time"

	"github.com/voxire/onion/internal/platform/money"
)

type BookingService struct {
	ID              int64
	OwnerID         int64
	BusinessID      int64
	Name            string
	Description     string
	DurationMinutes int
	Price           money.Amount
	Color           string
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type BookingServiceInput struct {
	Name            string
	Description     string
	DurationMinutes int
	Price           money.Amount
	Color           string
	IsActive        bool
}
