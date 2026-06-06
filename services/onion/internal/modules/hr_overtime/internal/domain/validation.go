package domain

import (
	"errors"
	"strings"
)

var (
	ErrInvalidOvertimeStatus = errors.New("invalid overtime request status")
	ErrNegativeExtraMinutes  = errors.New("planned extra minutes must be positive")
	ErrNegativeMultiplier    = errors.New("overtime multiplier must be positive")
	ErrEmptyOvertimeDate     = errors.New("overtime date is required")
)

// IsValidOvertimeStatus returns true when the status is recognised.
func IsValidOvertimeStatus(status string) bool {
	switch status {
	case OvertimeStatusPending, OvertimeStatusApproved, OvertimeStatusRejected:
		return true
	}
	return false
}

// ValidateOvertimeRequest checks that required overtime request fields are present.
func ValidateOvertimeRequest(req OvertimeRequestRecord) error {
	if strings.TrimSpace(req.Date) == "" {
		return ErrEmptyOvertimeDate
	}
	if req.PlannedExtraMinutes <= 0 {
		return ErrNegativeExtraMinutes
	}
	if !IsValidOvertimeStatus(req.Status) {
		return ErrInvalidOvertimeStatus
	}
	return nil
}

// ValidateOvertimeConfig checks that the config has valid multiplier and threshold values.
func ValidateOvertimeConfig(cfg OvertimeConfigRecord) error {
	if cfg.DefaultMultiplier <= 0 {
		return ErrNegativeMultiplier
	}
	return nil
}
