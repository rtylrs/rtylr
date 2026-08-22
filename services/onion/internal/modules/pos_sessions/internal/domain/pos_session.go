package domain

import (
	"errors"
	"time"
)

var ErrSessionNotFound = errors.New("session not found")

type Session struct {
	ID           int64      `json:"id"`
	OwnerID      int64      `json:"owner_id"`
	BusinessID   int64      `json:"business_id"`
	LocationID   *int64     `json:"location_id"`
	StaffID      *int64     `json:"staff_id"`
	OpeningFloat float64    `json:"opening_float"`
	ClosingFloat float64    `json:"closing_float"`
	TotalSales   float64    `json:"total_sales"`
	Status       string     `json:"status"`
	OpenedAt     time.Time  `json:"opened_at"`
	ClosedAt     *time.Time `json:"closed_at"`
	CreatedAt    time.Time  `json:"created_at"`
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
