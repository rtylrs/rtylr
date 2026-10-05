package domain

import "time"

type LeaveTypeRecord struct {
	ID                             int64
	OwnerID                        int64
	BusinessID                     int64
	Name                           string
	IsPaid                         bool
	RequiresApproval               bool
	AccrualRule                    string
	AccrualAmount                  float64
	MaxCarryoverDays               int32
	MinNoticeDays                  int32
	RequiresDocumentationAfterDays int32
	IsActive                       bool
	CreatedAt                      time.Time
	UpdatedAt                      time.Time
}

// LeaveTypePatch preserves field presence for the generated update contract.
// A nil field means "not supplied"; in particular, false and zero are valid
// values when their pointer is non-nil.
type LeaveTypePatch struct {
	Name                           *string
	IsPaid                         *bool
	RequiresApproval               *bool
	AccrualRule                    *string
	AccrualAmount                  *float64
	MaxCarryoverDays               *int32
	MinNoticeDays                  *int32
	RequiresDocumentationAfterDays *int32
	IsActive                       *bool
}

func (p LeaveTypePatch) Empty() bool {
	return p.Name == nil &&
		p.IsPaid == nil &&
		p.RequiresApproval == nil &&
		p.AccrualRule == nil &&
		p.AccrualAmount == nil &&
		p.MaxCarryoverDays == nil &&
		p.MinNoticeDays == nil &&
		p.RequiresDocumentationAfterDays == nil &&
		p.IsActive == nil
}

// ApplyLeaveTypePatch returns a copy of current with only supplied fields
// changed. It is also used by the service to validate the effective record
// before it is persisted.
func ApplyLeaveTypePatch(current LeaveTypeRecord, patch LeaveTypePatch) LeaveTypeRecord {
	updated := current
	if patch.Name != nil {
		updated.Name = *patch.Name
	}
	if patch.IsPaid != nil {
		updated.IsPaid = *patch.IsPaid
	}
	if patch.RequiresApproval != nil {
		updated.RequiresApproval = *patch.RequiresApproval
	}
	if patch.AccrualRule != nil {
		updated.AccrualRule = *patch.AccrualRule
	}
	if patch.AccrualAmount != nil {
		updated.AccrualAmount = *patch.AccrualAmount
	}
	if patch.MaxCarryoverDays != nil {
		updated.MaxCarryoverDays = *patch.MaxCarryoverDays
	}
	if patch.MinNoticeDays != nil {
		updated.MinNoticeDays = *patch.MinNoticeDays
	}
	if patch.RequiresDocumentationAfterDays != nil {
		updated.RequiresDocumentationAfterDays = *patch.RequiresDocumentationAfterDays
	}
	if patch.IsActive != nil {
		updated.IsActive = *patch.IsActive
	}
	return updated
}

type LeaveRequestRecord struct {
	ID                   int64
	OwnerID              int64
	BusinessID           int64
	EmployeeID           int64
	LeaveTypeID          int64
	StartDate            string
	EndDate              string
	TotalDays            float64
	Notes                string
	Status               string
	CurrentApprovalLevel int32
	RejectionReason      string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type LeaveBalanceRecord struct {
	ID               int64
	OwnerID          int64
	BusinessID       int64
	EmployeeID       int64
	LeaveTypeID      int64
	Year             int32
	EntitledDays     float64
	UsedDays         float64
	PendingDays      float64
	LastAccrualMonth int32
	CreatedAt        time.Time
	UpdatedAt        time.Time
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
