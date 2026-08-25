package domain

import "time"

type PayrollConfigRecord struct {
	ID              int64     `json:"id"`
	OwnerID         int64     `json:"owner_id"`
	BusinessID      int64     `json:"business_id"`
	PeriodType      string    `json:"period_type"`
	PayDayOfMonth   int32     `json:"pay_day_of_month"`
	OvertimeEnabled bool      `json:"overtime_enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PayrollRunRecord struct {
	ID              int64     `json:"id"`
	OwnerID         int64     `json:"owner_id"`
	BusinessID      int64     `json:"business_id"`
	PeriodStart     string    `json:"period_start"`
	PeriodEnd       string    `json:"period_end"`
	Status          string    `json:"status"`
	CreatedBy       int64     `json:"created_by"`
	ApprovedBy      int64     `json:"approved_by"`
	ExportedAt      time.Time `json:"exported_at"`
	TotalGrossCents int64     `json:"total_gross_cents"`
	TotalNetCents   int64     `json:"total_net_cents"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PayrollEntryRecord struct {
	ID               int64     `json:"id"`
	OwnerID          int64     `json:"owner_id"`
	BusinessID       int64     `json:"business_id"`
	PayrollRunID     int64     `json:"payroll_run_id"`
	EmployeeID       int64     `json:"employee_id"`
	HoursWorked      float64   `json:"hours_worked"`
	OvertimeHours    float64   `json:"overtime_hours"`
	BasePayCents     int64     `json:"base_pay_cents"`
	OvertimePayCents int64     `json:"overtime_pay_cents"`
	BonusesCents     int64     `json:"bonuses_cents"`
	DeductionsCents  int64     `json:"deductions_cents"`
	NetPayCents      int64     `json:"net_pay_cents"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type TimesheetSummary struct {
	EmployeeID   int64
	TotalMinutes int32
}

type OvertimeSummary struct {
	EmployeeID        int64
	TotalMinutes      int32
	WeightedMinutes   float64
	UnweightedMinutes float64
}

type EmployeePayInfo struct {
	EmployeeID         int64
	PayType            string
	HourlyRateCents    int64
	MonthlySalaryCents int64
}

type ListPayrollRunsParams struct {
	Limit  int32
	Offset int32
}

type ApprovalChainInfo struct {
	TotalLevels int32
	Levels      []ApprovalLevelInfo
}

type ApprovalLevelInfo struct {
	LevelOrder         int32
	ApproverType       string
	ApproverRoleID     int64
	ApproverEmployeeID int64
}
