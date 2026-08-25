package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrPayrollNotFound = errors.New("payroll record not found")

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

func (p *PayrollRecord) FromRow(row *sql.Rows) error {
	var (
		regularHours  sql.NullFloat64
		overtimeHours sql.NullFloat64
		overtimePay   sql.NullFloat64
		bonuses       sql.NullFloat64
		deductions    sql.NullFloat64
		taxWithheld   sql.NullFloat64
		status        sql.NullString
		paidAt        sql.NullTime
		notes         sql.NullString
	)

	err := row.Scan(
		&p.ID,
		&p.OwnerID,
		&p.BusinessID,
		&p.StaffID,
		&p.PeriodStart,
		&p.PeriodEnd,
		&p.PeriodLabel,
		&regularHours,
		&overtimeHours,
		&p.BasePay,
		&overtimePay,
		&bonuses,
		&deductions,
		&taxWithheld,
		&p.NetPay,
		&status,
		&paidAt,
		&notes,
		&p.CreatedAt,
	)
	if err != nil {
		return err
	}

	if regularHours.Valid {
		p.RegularHours = regularHours.Float64
	}

	if overtimeHours.Valid {
		p.OvertimeHours = overtimeHours.Float64
	}

	if overtimePay.Valid {
		p.OvertimePay = overtimePay.Float64
	}

	if bonuses.Valid {
		p.Bonuses = bonuses.Float64
	}

	if deductions.Valid {
		p.Deductions = deductions.Float64
	}

	if taxWithheld.Valid {
		p.TaxWithheld = taxWithheld.Float64
	}

	if status.Valid {
		p.Status = status.String
	}

	if paidAt.Valid {
		p.PaidAt = paidAt.Time
	}

	if notes.Valid {
		p.Notes = notes.String
	}

	return nil
}

type PayrollRunInput struct {
	OwnerID     int64  `json:"owner_id"`
	BusinessID  int64  `json:"business_id"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
	PeriodLabel string `json:"period_label"`
}

type PayrollMarkPaidInput struct {
	ID int64 `json:"id"`
}

type DashboardStats struct {
	TotalStaff     int `json:"total_staff"`
	OnShiftToday   int `json:"on_shift_today"`
	PendingTimeOff int `json:"pending_time_off"`
	PayrollDraft   int `json:"payroll_draft"`
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

// CentsToFloat converts integer cents to standard float amount.
func CentsToFloat(cents int64) float64 {
	return float64(cents) / 100.0
}
