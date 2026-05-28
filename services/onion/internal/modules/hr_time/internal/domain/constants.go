package domain

// Time entry status values.
const (
	EntryStatusDraft    = "draft"
	EntryStatusApproved = "approved"
	EntryStatusRejected = "rejected"
)

// Timesheet status values.
const (
	TimesheetStatusDraft     = "draft"
	TimesheetStatusSubmitted = "submitted"
	TimesheetStatusApproved  = "approved"
	TimesheetStatusRejected  = "rejected"
)

// Time entry source values.
const (
	EntrySourceManual = "manual"
	EntrySourceFlow   = "flow"
	EntrySourceImport = "import"
)

// Approval action values.
const (
	ApprovalActionApprove = "approve"
	ApprovalActionReject  = "reject"
)
