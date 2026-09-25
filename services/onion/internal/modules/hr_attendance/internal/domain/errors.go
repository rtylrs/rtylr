package domain

import "errors"

var (
	ErrAlreadyClockedIn      = errors.New("employee is already clocked in")
	ErrNotClockedIn          = errors.New("employee is not clocked in")
	ErrInvalidQRToken        = errors.New("invalid or expired QR token")
	ErrGeofenceViolation     = errors.New("clock-in location is outside the allowed geofence")
	ErrClockInMethodDisabled = errors.New("this clock-in method is not enabled")
	ErrAttendanceNotFound    = errors.New("attendance record not found")
	ErrInvalidPIN            = errors.New("invalid clock-in PIN")
	ErrPINNotSet             = errors.New("employee has not set a clock-in PIN")
)
