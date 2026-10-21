package domain

import (
	"errors"
	"time"
)

var ErrSessionNotFound = errors.New("session not found")

type Session struct {
	ID           int64
	OwnerID      int64
	BusinessID   int64
	LocationID   *int64
	StaffID      *int64
	OpeningFloat float64
	ClosingFloat float64
	TotalSales   float64
	Status       string
	OpenedAt     time.Time
	ClosedAt     *time.Time
	CreatedAt    time.Time
}

type OpenSessionInput struct {
	OwnerID      int64
	BusinessID   int64
	LocationID   *int64
	StaffID      *int64
	OpeningFloat float64
}

type CloseSessionInput struct {
	ID           int64
	OwnerID      int64
	BusinessID   int64
	StaffID      int64
	ClosingFloat float64
}
