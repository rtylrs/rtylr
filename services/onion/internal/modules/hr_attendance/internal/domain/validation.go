package domain

import "errors"

var (
	ErrInvalidClockInMethod    = errors.New("invalid clock-in method")
	ErrInvalidAttendanceStatus = errors.New("invalid attendance status")
	ErrNegativeDuration        = errors.New("duration minutes must be non-negative")
)

// IsValidClockInMethod returns true when the method is recognised.
func IsValidClockInMethod(method string) bool {
	switch method {
	case ClockInMethodPin, ClockInMethodQR, ClockInMethodApp, ClockInMethodFacial, ClockInMethodGPS:
		return true
	}
	return false
}

// IsValidAttendanceStatus returns true when the status is recognised.
func IsValidAttendanceStatus(status string) bool {
	switch status {
	case AttendanceStatusClockedIn, AttendanceStatusClockedOut, AttendanceStatusAbsent:
		return true
	}
	return false
}

// ValidateAttendanceRecord checks required fields on an attendance record.
func ValidateAttendanceRecord(rec AttendanceRecordRecord) error {
	if !IsValidClockInMethod(rec.ClockInMethod) {
		return ErrInvalidClockInMethod
	}
	if !IsValidAttendanceStatus(rec.Status) {
		return ErrInvalidAttendanceStatus
	}
	if rec.DurationMinutes < 0 {
		return ErrNegativeDuration
	}
	return nil
}
