package domain

// Time entry status values.
const (
	EntryStatusDraft    = "draft"
	EntryStatusApproved = "approved"
	EntryStatusRejected = "rejected"
)

// Existing hr_core role permission keys used by hr_time. manage_employees is
// the current broad permission for HR staff who may administer another
// employee's record; approve_time_off is also the product's timesheet approval
// permission.
const (
	PermissionManageEmployees = "manage_employees"
	PermissionApproveTimeOff  = "approve_time_off"
)

// Timesheet status values.
const (
	TimesheetStatusDraft     = "draft"
	TimesheetStatusSubmitted = "submitted"
	TimesheetStatusApproved  = "approved"
	TimesheetStatusRejected  = "rejected"
	TimesheetStatusLocked    = "locked"
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
