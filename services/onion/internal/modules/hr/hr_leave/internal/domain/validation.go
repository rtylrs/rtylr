package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrEmptyLeaveTypeName = errors.New("leave type name is required")
	ErrNegativeAccrual    = errors.New("accrual amount must be non-negative")
	ErrInvalidAccrualRule = errors.New("invalid accrual rule")
)

// ValidateLeaveType checks that a leave type record has valid fields.
func ValidateLeaveType(lt LeaveTypeRecord) error {
	if strings.TrimSpace(lt.Name) == "" {
		return ErrEmptyLeaveTypeName
	}
	if lt.AccrualAmount < 0 {
		return ErrNegativeAccrual
	}
	if lt.MaxCarryoverDays < 0 || lt.MinNoticeDays < 0 || lt.RequiresDocumentationAfterDays < 0 {
		return ErrNegativeLeavePolicyLimit
	}
	rule := strings.ToLower(strings.TrimSpace(lt.AccrualRule))
	if rule != "" && !IsValidAccrualRule(rule) {
		return ErrInvalidAccrualRule
	}
	return nil
}

// ValidateDateRange returns an error when endDate is before startDate.
// Dates must be in "2006-01-02" format.
func ValidateDateRange(startDate, endDate string) error {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return ErrInvalidDateRange
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return ErrInvalidDateRange
	}
	if end.Before(start) {
		return ErrInvalidDateRange
	}
	return nil
}

// IsValidLeaveStatus returns true when the status is a recognised value.
func IsValidLeaveStatus(status string) bool {
	switch status {
	case LeaveStatusPending, LeaveStatusApproved, LeaveStatusRejected, LeaveStatusCanceled:
		return true
	}
	return false
}

// IsValidAccrualRule returns true when the rule is a recognised value.
func IsValidAccrualRule(rule string) bool {
	switch rule {
	case AccrualRuleNone, AccrualRuleMonthly, AccrualRuleAnnual:
		return true
	}
	return false
}

// RemainingBalance computes the remaining leave balance.
func RemainingBalance(bal LeaveBalanceRecord) float64 {
	return bal.EntitledDays - bal.UsedDays - bal.PendingDays
}
