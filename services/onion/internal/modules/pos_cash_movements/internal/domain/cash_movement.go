package domain

import (
	"time"
)

type CashMovement struct {
	ID         int64     `json:"id"`
	BusinessID int64     `json:"business_id"`
	BranchID   int64     `json:"branch_id"`
	TerminalID *int64    `json:"terminal_id"`
	SessionID  int64     `json:"session_id"`
	EmployeeID int64     `json:"employee_id"`
	Type       string    `json:"type"`
	Amount     float64   `json:"amount"`
	Reason     *string   `json:"reason"`
	ApprovedBy *int64    `json:"approved_by"`
	CreatedAt  time.Time `json:"created_at"`
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

// CentsToFloat converts integer cents to standard float amount.
func CentsToFloat(cents int64) float64 {
	return float64(cents) / 100.0
}
