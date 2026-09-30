package domain

import (
	"time"
)

type PayrollRecord struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"owner_id"`
	BusinessID    int64     `json:"business_id"`
	StaffID       int64     `json:"staff_id"`
	PeriodStart   time.Time `json:"period_start"`
	PeriodEnd     time.Time `json:"period_end"`
	PeriodLabel   string    `json:"period_label"`
	RegularHours  float64   `json:"regular_hours"`
	OvertimeHours float64   `json:"overtime_hours"`
	BasePay       float64   `json:"base_pay"`
	OvertimePay   float64   `json:"overtime_pay"`
	Bonuses       float64   `json:"bonuses"`
	Deductions    float64   `json:"deductions"`
	TaxWithheld   float64   `json:"tax_withheld"`
	NetPay        float64   `json:"net_pay"`
	Status        string    `json:"status"`
	PaidAt        time.Time `json:"paid_at"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
}

// StatusPaid is the only status that means money has left the business.
const StatusPaid = "paid"

// PeriodRow is the part of a payroll record a period-wide decision needs. Both
// re-running a period and settling one record read the whole period under a
// lock, and both decisions are about paid rows.
type PeriodRow struct {
	ID      int64
	StaffID int64
	Status  string
}

// PeriodHasPaidRecord reports whether any row in the period has been settled.
// Re-running such a period is refused: recomputing it would restate figures for
// money that has already been paid out.
func PeriodHasPaidRecord(rows []PeriodRow) bool {
	for _, row := range rows {
		if row.Status == StatusPaid {
			return true
		}
	}
	return false
}

// StaffHasPaidRecordInPeriod reports whether the given employee already has a
// settled record for this period, ignoring one record id. That is the state a
// duplicated payroll run leaves behind, and settling the second row would pay
// the employee twice for one period. Records for other employees are the
// normal case of paying a period out employee by employee, so they do not
// block anyone.
func StaffHasPaidRecordInPeriod(rows []PeriodRow, staffID, exceptID int64) bool {
	for _, row := range rows {
		if row.ID != exceptID && row.StaffID == staffID && row.Status == StatusPaid {
			return true
		}
	}
	return false
}

type PayrollRunInput struct {
	OwnerID     int64  `json:"owner_id"`
	BusinessID  int64  `json:"business_id"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
	PeriodLabel string `json:"period_label"`
}

type PayrollRecordInput struct {
	StaffID       int64
	PeriodStart   time.Time
	PeriodEnd     time.Time
	PeriodLabel   string
	RegularHours  float64
	OvertimeHours float64
	BasePay       float64
	OvertimePay   float64
	NetPay        float64
}

// ShiftSummary is used internally for payroll calculation aggregation.
type ShiftSummary struct {
	StaffID            int64
	TotalHours         float64
	TotalOvertimeHours float64
}

// StaffPayRates holds the pay rate fields needed for payroll calculation.
type StaffPayRates struct {
	Salary       float64
	HourlyRate   float64
	OvertimeRate float64
}
