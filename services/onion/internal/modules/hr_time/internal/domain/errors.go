package domain

import "errors"

var (
	ErrTimesheetNotFound          = errors.New("timesheet not found")
	ErrTimesheetAlreadySubmitted  = errors.New("timesheet already submitted")
	ErrTimesheetLocked            = errors.New("timesheet is locked")
	ErrTimeEntryNotFound          = errors.New("time entry not found")
	ErrApprovalChainNotConfigured = errors.New("approval chain not configured for timesheets")
	ErrInvalidDuration            = errors.New("duration must be greater than 0")
	ErrDateInLockedPeriod         = errors.New("cannot add time entry to a locked period")
)
