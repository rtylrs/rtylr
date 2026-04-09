package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrPayrollNotFound = errors.New("payroll record not found")

type PayrollRecord struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"ownerId"`
	StaffID      int64     `json:"staffId"`
	PeriodStart  time.Time `json:"periodStart"`
	PeriodEnd    time.Time `json:"periodEnd"`
	PeriodLabel  string    `json:"periodLabel"`
	RegularHours float64   `json:"regularHours"`
	OvertimeHours float64  `json:"overtimeHours"`
	BasePay      float64   `json:"basePay"`
	OvertimePay  float64   `json:"overtimePay"`
	Bonuses      float64   `json:"bonuses"`
	Deductions   float64   `json:"deductions"`
	TaxWithheld  float64   `json:"taxWithheld"`
	NetPay       float64   `json:"netPay"`
	Status       string    `json:"status"`
	PaidAt       time.Time `json:"paidAt"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"createdAt"`
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
	OwnerID     int64  `json:"ownerId"`
	PeriodStart string `json:"periodStart"`
	PeriodEnd   string `json:"periodEnd"`
	PeriodLabel string `json:"periodLabel"`
}

type PayrollMarkPaidInput struct {
	ID int64 `json:"id"`
}

type DashboardStats struct {
	TotalStaff       int `json:"totalStaff"`
	OnShiftToday     int `json:"onShiftToday"`
	PendingTimeOff   int `json:"pendingTimeOff"`
	PayrollDraft     int `json:"payrollDraft"`
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
