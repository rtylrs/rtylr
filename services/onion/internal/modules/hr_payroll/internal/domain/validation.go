package domain

import (
	"errors"
	"strings"
)

var (
	ErrInvalidPayrollStatus = errors.New("invalid payroll run status")
	ErrInvalidPeriodType    = errors.New("invalid payroll period type")
	ErrNegativePayAmount    = errors.New("pay amounts must be non-negative")
	ErrEmptyPeriodDates     = errors.New("period start and end dates are required")
)

// IsValidPayrollStatus returns true when the status is recognised.
func IsValidPayrollStatus(status string) bool {
	switch status {
	case PayrollStatusDraft, PayrollStatusInReview, PayrollStatusApproved, PayrollStatusExported:
		return true
	}
	return false
}

// IsValidPeriodType returns true when the period type is recognised.
func IsValidPeriodType(periodType string) bool {
	switch periodType {
	case PayrollPeriodMonthly, PayrollPeriodBiWeekly, PayrollPeriodWeekly, PayrollPeriodSemiMonthly:
		return true
	}
	return false
}

// ValidatePayrollRun checks that required payroll run fields are present.
func ValidatePayrollRun(run PayrollRunRecord) error {
	if strings.TrimSpace(run.PeriodStart) == "" || strings.TrimSpace(run.PeriodEnd) == "" {
		return ErrEmptyPeriodDates
	}
	if !IsValidPayrollStatus(run.Status) {
		return ErrInvalidPayrollStatus
	}
	if run.TotalGrossCents < 0 || run.TotalNetCents < 0 {
		return ErrNegativePayAmount
	}
	return nil
}

// ValidatePayrollEntry checks that a payroll entry has valid amounts.
func ValidatePayrollEntry(entry PayrollEntryRecord) error {
	if entry.BasePayCents < 0 || entry.OvertimePayCents < 0 || entry.BonusesCents < 0 || entry.DeductionsCents < 0 || entry.NetPayCents < 0 {
		return ErrNegativePayAmount
	}
	return nil
}

// ComputeNetPay derives net pay, in cents, from the entry components.
func ComputeNetPay(entry PayrollEntryRecord) int64 {
	return entry.BasePayCents + entry.OvertimePayCents + entry.BonusesCents - entry.DeductionsCents
}
