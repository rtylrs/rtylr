package domain

import (
	"time"
)

type CashMovement struct {
	ID         int64
	BusinessID int64
	BranchID   int64
	TerminalID *int64
	SessionID  int64
	EmployeeID int64
	Type       string
	Amount     float64
	Reason     *string
	ApprovedBy *int64
	CreatedAt  time.Time
}

type CreateCashMovementInput struct {
	BusinessID int64
	BranchID   int64
	TerminalID *int64
	SessionID  int64
	EmployeeID int64
	Type       string
	Amount     float64
	Reason     *string
	ApprovedBy *int64
}
