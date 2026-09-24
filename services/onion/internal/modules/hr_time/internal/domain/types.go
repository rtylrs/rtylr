package domain

import "time"

type TimeEntryRecord struct {
	ID                int64     `json:"id"`
	OwnerID           int64     `json:"owner_id"`
	BusinessID        int64     `json:"business_id"`
	EmployeeID        int64     `json:"employee_id"`
	TimesheetID       int64     `json:"timesheet_id"`
	ProjectID         int64     `json:"project_id"`
	TaskID            int64     `json:"task_id"`
	CategoryID        int64     `json:"category_id"`
	Date              string    `json:"date"`
	DurationMinutes   int32     `json:"duration_minutes"`
	StartTime         string    `json:"start_time"`
	EndTime           string    `json:"end_time"`
	Description       string    `json:"description"`
	IsBillable        bool      `json:"is_billable"`
	Status            string    `json:"status"`
	RejectionReason   string    `json:"rejection_reason"`
	ApprovedBy        int64     `json:"approved_by"`
	Source            string    `json:"source"`
	BillableInvoiceID int64     `json:"billable_invoice_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
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
	ID                   int64     `json:"id"`
	OwnerID              int64     `json:"owner_id"`
	BusinessID           int64     `json:"business_id"`
	EmployeeID           int64     `json:"employee_id"`
	PeriodStart          string    `json:"period_start"`
	PeriodEnd            string    `json:"period_end"`
	Status               string    `json:"status"`
	CurrentApprovalLevel int32     `json:"current_approval_level"`
	SubmittedAt          time.Time `json:"submitted_at"`
	TotalMinutes         int32     `json:"total_minutes"`
	RejectionReason      string    `json:"rejection_reason"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type ApprovalActionRecord struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	BusinessID  int64     `json:"business_id"`
	TimesheetID int64     `json:"timesheet_id"`
	ApproverID  int64     `json:"approver_id"`
	Action      string    `json:"action"`
	Level       int32     `json:"level"`
	Comment     string    `json:"comment"`
	CreatedAt   time.Time `json:"created_at"`
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

type BillableExportItem struct {
	EntryID         int64  `json:"entry_id"`
	EmployeeID      int64  `json:"employee_id"`
	ProjectID       int64  `json:"project_id"`
	TaskID          int64  `json:"task_id"`
	Date            string `json:"date"`
	DurationMinutes int32  `json:"duration_minutes"`
	Description     string `json:"description"`
	IsBillable      bool   `json:"is_billable"`
}
