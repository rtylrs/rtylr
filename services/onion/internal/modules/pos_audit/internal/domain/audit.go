package domain

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID         int64            `json:"id"`
	BusinessID int64            `json:"business_id"`
	BranchID   int64            `json:"branch_id"`
	TerminalID *int64           `json:"terminal_id"`
	EmployeeID int64            `json:"employee_id"`
	Action     string           `json:"action"`
	EntityType string           `json:"entity_type"`
	EntityID   *int64           `json:"entity_id"`
	Details    *json.RawMessage `json:"details"`
	IPAddress  *string          `json:"ip_address"`
	CreatedAt  time.Time        `json:"created_at"`
}

type AuditLogInput struct {
	BusinessID int64
	BranchID   int64
	TerminalID *int64
	EmployeeID int64
	Action     string
	EntityType string
	EntityID   *int64
	Details    *json.RawMessage
	IPAddress  *string
}

type AuditLogFilter struct {
	BusinessID      int64
	Action          *string
	EntityType      *string
	EmployeeID      *int64
	StartDate       *string
	EndDate         *string
	Limit           int32
	CursorCreatedAt *time.Time
	CursorID        *int64
}

type AuditLogPage struct {
	Logs       []AuditLog
	NextCursor string
}
