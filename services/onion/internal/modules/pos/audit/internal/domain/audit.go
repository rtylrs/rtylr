package domain

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID         int64
	BusinessID int64
	BranchID   int64
	TerminalID *int64
	EmployeeID int64
	Action     string
	EntityType string
	EntityID   *int64
	Details    *json.RawMessage
	IPAddress  *string
	CreatedAt  time.Time
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
