package domain

import (
	"errors"
	"time"
)

var ErrTerminalNotFound = errors.New("terminal not found")

type Terminal struct {
	ID         int64      `json:"id"`
	BusinessID int64      `json:"businessId"`
	BranchID   int64      `json:"branchId"`
	Name       string     `json:"name"`
	DeviceID   *string    `json:"deviceId"`
	IsActive   bool       `json:"isActive"`
	LastSeenAt *time.Time `json:"lastSeenAt"`
	CreatedAt  time.Time  `json:"createdAt"`
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
