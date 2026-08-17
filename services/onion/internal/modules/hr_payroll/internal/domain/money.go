package domain

import "math"

// Payroll amounts (PayrollRunRecord.TotalGrossCents/TotalNetCents,
// PayrollEntryRecord.BasePayCents/OvertimePayCents/BonusesCents/
// DeductionsCents/NetPayCents, EmployeePayInfo.HourlyRateCents/
// MonthlySalaryCents) are stored internally as integer minor units (cents)
// so that persistence and repeated read/write cycles never lose or drift
// precision the way float64 dollars silently can.
//
// The HTTP, Caret-tool, and cross-module finance journal contracts still
// carry these amounts as float64 dollars on the wire (matching the
// generated hr/employees/payroll/runs/v1 proto and the financejournal
// platform contract, both owned outside this module and out of scope for
// this fix), so every boundary that reads or writes those contracts must go
// through these two conversions exactly once - never propagate a float64
// dollar amount through further arithmetic before it has been converted to
// cents.

// DollarsToCents converts a float64 dollar amount from an external contract
// into integer cents, rounding to the nearest cent.
func DollarsToCents(dollars float64) int64 {
	return int64(math.Round(dollars * 100))
}

// CentsToDollars converts an internally stored integer cents amount back
// into the float64 dollars shape required by an external contract.
func CentsToDollars(cents int64) float64 {
	return float64(cents) / 100
}
