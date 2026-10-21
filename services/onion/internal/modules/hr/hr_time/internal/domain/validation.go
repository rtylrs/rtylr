package domain

import (
	"errors"
	"strings"
)

var (
	ErrInvalidEntryStatus     = errors.New("invalid time entry status")
	ErrInvalidTimesheetStatus = errors.New("invalid timesheet status")
	ErrNegativeEntryDuration  = errors.New("time entry duration must be non-negative")
	ErrEmptyEntryDate         = errors.New("time entry date is required")
	ErrEmptyPeriodDates       = errors.New("timesheet period start and end dates are required")
)

// IsValidEntryStatus returns true when the status is recognised.
func IsValidEntryStatus(status string) bool {
	switch status {
	case EntryStatusDraft, EntryStatusApproved, EntryStatusRejected:
		return true
	}
	return false
}

// IsValidTimesheetStatus returns true when the status is recognised.
func IsValidTimesheetStatus(status string) bool {
	switch status {
	case TimesheetStatusDraft, TimesheetStatusSubmitted, TimesheetStatusApproved, TimesheetStatusRejected:
		return true
	}
	return false
}

// IsValidEntrySource returns true when the source is recognised.
func IsValidEntrySource(source string) bool {
	switch source {
	case EntrySourceManual, EntrySourceFlow, EntrySourceImport:
		return true
	}
	return false
}

// ValidateTimeEntry checks required fields on a time entry.
func ValidateTimeEntry(e TimeEntryRecord) error {
	if strings.TrimSpace(e.Date) == "" {
		return ErrEmptyEntryDate
	}
	if e.DurationMinutes < 0 {
		return ErrNegativeEntryDuration
	}
	if e.Status != "" && !IsValidEntryStatus(e.Status) {
		return ErrInvalidEntryStatus
	}
	return nil
}

// ValidateTimesheet checks required fields on a timesheet record.
func ValidateTimesheet(ts TimesheetRecord) error {
	if strings.TrimSpace(ts.PeriodStart) == "" || strings.TrimSpace(ts.PeriodEnd) == "" {
		return ErrEmptyPeriodDates
	}
	if ts.Status != "" && !IsValidTimesheetStatus(ts.Status) {
		return ErrInvalidTimesheetStatus
	}
	return nil
}
