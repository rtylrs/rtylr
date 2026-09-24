package domain

import "errors"

var (
	ErrTimesheetNotFound           = errors.New("timesheet not found")
	ErrTimesheetAlreadySubmitted   = errors.New("timesheet already submitted")
	ErrTimesheetLocked             = errors.New("timesheet is locked")
	ErrTimesheetNotPending         = errors.New("timesheet is not awaiting approval")
	ErrTimesheetAccessDenied       = errors.New("timesheet access denied")
	ErrTimesheetApproverNotAllowed = errors.New("approver is not allowed by the timesheet approval chain")
	ErrTimesheetSelfApproval       = errors.New("employees cannot approve their own timesheets")
	ErrTimeEntryNotFound           = errors.New("time entry not found")
	ErrTimeEntryAccessDenied       = errors.New("time entry access denied")
	ErrInvoiceNotFound             = errors.New("invoice not found")
	ErrApprovalChainNotConfigured  = errors.New("approval chain not configured for timesheets")
	ErrInvalidDuration             = errors.New("duration must be greater than 0")
	ErrDateInLockedPeriod          = errors.New("cannot add time entry to a locked period")
)
