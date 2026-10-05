package domain

import "time"

type TimeEntryRecord struct {
	ID                int64
	OwnerID           int64
	BusinessID        int64
	EmployeeID        int64
	TimesheetID       int64
	ProjectID         int64
	TaskID            int64
	CategoryID        int64
	Date              string
	DurationMinutes   int32
	StartTime         string
	EndTime           string
	Description       string
	IsBillable        bool
	Status            string
	RejectionReason   string
	ApprovedBy        int64
	Source            string
	BillableInvoiceID int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// TimeEntryPatch carries only fields supplied by an update request. A nil
// field preserves the stored value; workflow-owned fields are intentionally
// absent so an ordinary update cannot clear approval state or ownership links.
type TimeEntryPatch struct {
	EmployeeID      *int64
	ProjectID       *int64
	TaskID          *int64
	CategoryID      *int64
	Date            *string
	DurationMinutes *int32
	StartTime       *string
	EndTime         *string
	Description     *string
	IsBillable      *bool
	Source          *string
}

type TimesheetRecord struct {
	ID                   int64
	OwnerID              int64
	BusinessID           int64
	EmployeeID           int64
	PeriodStart          string
	PeriodEnd            string
	Status               string
	CurrentApprovalLevel int32
	SubmittedAt          time.Time
	TotalMinutes         int32
	RejectionReason      string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type ApprovalActionRecord struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	TimesheetID int64
	ApproverID  int64
	Action      string
	Level       int32
	Comment     string
	CreatedAt   time.Time
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

// ListTimeEntriesParams selects logged time. Every filter is optional and they
// combine with AND, matching hr.time_entry.v1 ListTimeEntriesRequest.
//
// The optional fields are pointers rather than zero values because zero is a
// legitimate answer for none of them and an ambiguous one for all: employee 0
// does not exist, but "" is not a date and false is a real billability. A zero
// sentinel here is what made employee_id a mandatory filter before
// tryrtylr/rtylr#855.
type ListTimeEntriesParams struct {
	EmployeeID *int64
	FromDate   *string
	ToDate     *string
	ProjectID  *int64
	IsBillable *bool
	Status     *string
	Limit      int32
	Offset     int32
}

// ListTimesheetsParams selects timesheets for the approvals screen. Both
// filters are optional and combine with AND, matching the contract's
// hr.time_entry.v1 ListTimesheetsRequest ("absent to include all
// employees/all statuses"). They are pointers so an absent filter is not
// confused with a zero one: the previous int64 employee_id kept the zero
// sentinel and made absent mean employee 0, so a status-only query always
// came back empty.
type ListTimesheetsParams struct {
	EmployeeID *int64
	Status     *string
	Limit      int32
	Offset     int32
}
