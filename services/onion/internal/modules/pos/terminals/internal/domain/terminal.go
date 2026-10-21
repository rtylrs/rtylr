package domain

import (
	"errors"
	"time"
)

var ErrTerminalNotFound = errors.New("terminal not found")

type Terminal struct {
	ID         int64
	BusinessID int64
	BranchID   int64
	Name       string
	DeviceID   *string
	IsActive   bool
	LastSeenAt *time.Time
	CreatedAt  time.Time
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
