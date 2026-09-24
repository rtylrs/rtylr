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

// ValidateOvertimeDecision enforces the reason requirement at the domain
// boundary as well as at the HTTP adapter. Approved decisions do not need a
// rejection reason; rejected decisions must carry a non-blank explanation.
func ValidateOvertimeDecision(approved bool, rejectionReason string) error {
	if !approved && strings.TrimSpace(rejectionReason) == "" {
		return ErrOvertimeRejectionReason
	}
	return nil
}

// NormalizeOvertimeListParams canonicalizes optional filters before they reach
// the repository or cache. Empty/whitespace status is treated as absent.
func NormalizeOvertimeListParams(params ListOvertimeRequestsParams) ListOvertimeRequestsParams {
	if params.EmployeeID != nil {
		employeeID := *params.EmployeeID
		params.EmployeeID = &employeeID
	}
	if params.Status != nil {
		status := strings.ToLower(strings.TrimSpace(*params.Status))
		if status == "" {
			params.Status = nil
		} else {
			params.Status = &status
		}
	}
	if params.Offset < 0 {
		params.Offset = 0
	}
	return params
}
