package domain

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID         int64            `json:"id"`
	BusinessID int64            `json:"businessId"`
	BranchID   int64            `json:"branchId"`
	TerminalID *int64           `json:"terminalId"`
	EmployeeID int64            `json:"employeeId"`
	Action     string           `json:"action"`
	EntityType string           `json:"entityType"`
	EntityID   *int64           `json:"entityId"`
	Details    *json.RawMessage `json:"details"`
	IPAddress  *string          `json:"ipAddress"`
	CreatedAt  time.Time        `json:"createdAt"`
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
	BusinessID int64
	Action     *string
	EntityType *string
	EmployeeID *int64
	StartDate  *string
	EndDate    *string
}
