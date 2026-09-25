package domain

import (
	"time"
)

type Terminal struct {
	ID         int64      `json:"id"`
	BusinessID int64      `json:"business_id"`
	BranchID   int64      `json:"branch_id"`
	Name       string     `json:"name"`
	DeviceID   *string    `json:"device_id"`
	IsActive   bool       `json:"is_active"`
	LastSeenAt *time.Time `json:"last_seen_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

type TerminalInput struct {
	BusinessID int64
	BranchID   int64
	Name       string
	DeviceID   *string
}

type TerminalUpdateInput struct {
	ID         int64
	BusinessID int64
	BranchID   int64
	Name       string
	DeviceID   *string
	IsActive   bool
}
