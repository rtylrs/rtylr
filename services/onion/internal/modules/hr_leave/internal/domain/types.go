package domain

import "time"

type LeaveTypeRecord struct {
	ID                             int64     `json:"id"`
	OwnerID                        int64     `json:"owner_id"`
	BusinessID                     int64     `json:"business_id"`
	Name                           string    `json:"name"`
	IsPaid                         bool      `json:"is_paid"`
	RequiresApproval               bool      `json:"requires_approval"`
	AccrualRule                    string    `json:"accrual_rule"`
	AccrualAmount                  float64   `json:"accrual_amount"`
	MaxCarryoverDays               int32     `json:"max_carryover_days"`
	MinNoticeDays                  int32     `json:"min_notice_days"`
	RequiresDocumentationAfterDays int32     `json:"requires_documentation_after_days"`
	IsActive                       bool      `json:"is_active"`
	CreatedAt                      time.Time `json:"created_at"`
	UpdatedAt                      time.Time `json:"updated_at"`
}

type LeaveRequestRecord struct {
	ID                   int64     `json:"id"`
	OwnerID              int64     `json:"owner_id"`
	BusinessID           int64     `json:"business_id"`
	EmployeeID           int64     `json:"employee_id"`
	LeaveTypeID          int64     `json:"leave_type_id"`
	StartDate            string    `json:"start_date"`
	EndDate              string    `json:"end_date"`
	TotalDays            float64   `json:"total_days"`
	Notes                string    `json:"notes"`
	Status               string    `json:"status"`
	CurrentApprovalLevel int32     `json:"current_approval_level"`
	RejectionReason      string    `json:"rejection_reason"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type LeaveBalanceRecord struct {
	ID               int64     `json:"id"`
	OwnerID          int64     `json:"owner_id"`
	BusinessID       int64     `json:"business_id"`
	EmployeeID       int64     `json:"employee_id"`
	LeaveTypeID      int64     `json:"leave_type_id"`
	Year             int32     `json:"year"`
	EntitledDays     float64   `json:"entitled_days"`
	UsedDays         float64   `json:"used_days"`
	PendingDays      float64   `json:"pending_days"`
	LastAccrualMonth int32     `json:"last_accrual_month"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// LeaveDayAllocation splits a request across calendar-year balance buckets.
// A request crossing New Year's Eve must debit and restore each year's bucket
// independently rather than charging everything to the start year.
type LeaveDayAllocation struct {
	Year int32
	Days float64
}

type ListLeaveTypesParams struct {
	Limit  int32
	Offset int32
}

type ListLeaveRequestsParams struct {
	EmployeeID int64
	Status     string
	// FromDate/ToDate bound the overlap window (ISO-8601 dates): a request is
	// included when it starts on/before ToDate and ends on/after FromDate.
	// Nil leaves that side of the range open.
	FromDate *string
	ToDate   *string
	Limit    int32
	Offset   int32
}

type LeaveCalendarParams struct {
	LocationID int64
	DateFrom   string
	DateTo     string
}

type ApprovalChainInfo struct {
	TotalLevels int32
	Levels      []ApprovalLevelInfo
}

type ApprovalLevelInfo struct {
	LevelOrder         int32
	ApproverType       string
	ApproverRoleID     int64
	ApproverEmployeeID int64
}
