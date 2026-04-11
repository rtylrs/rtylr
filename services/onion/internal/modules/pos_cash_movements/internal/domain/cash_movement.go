package domain

import (
	"errors"
	"time"
)

var ErrCashMovementNotFound = errors.New("cash movement not found")

type CashMovement struct {
	ID         int64     `json:"id"`
	BusinessID int64     `json:"businessId"`
	BranchID   int64     `json:"branchId"`
	TerminalID *int64    `json:"terminalId"`
	SessionID  int64     `json:"sessionId"`
	EmployeeID int64     `json:"employeeId"`
	Type       string    `json:"type"`
	Amount     float64   `json:"amount"`
	Reason     *string   `json:"reason"`
	ApprovedBy *int64    `json:"approvedBy"`
	CreatedAt  time.Time `json:"createdAt"`
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
