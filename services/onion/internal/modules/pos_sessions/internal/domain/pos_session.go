package domain

import (
	"errors"
	"time"
)

var ErrSessionNotFound = errors.New("session not found")

type Session struct {
	ID           int64      `json:"id"`
	OwnerID      int64      `json:"ownerId"`
	LocationID   *int64     `json:"locationId"`
	StaffID      *int64     `json:"staffId"`
	OpeningFloat float64    `json:"openingFloat"`
	ClosingFloat float64    `json:"closingFloat"`
	TotalSales   float64    `json:"totalSales"`
	Status       string     `json:"status"`
	OpenedAt     time.Time  `json:"openedAt"`
	ClosedAt     *time.Time `json:"closedAt"`
	CreatedAt    time.Time  `json:"createdAt"`
}

type OpenSessionInput struct {
	OwnerID      int64
	LocationID   *int64
	StaffID      *int64
	OpeningFloat float64
}

type CloseSessionInput struct {
	ID           int64
	OwnerID      int64
	ClosingFloat float64
}
