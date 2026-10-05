package domain

import "time"

type PayrollConfigRecord struct {
	ID              int64
	OwnerID         int64
	BusinessID      int64
	PeriodType      string
	PayDayOfMonth   int32
	OvertimeEnabled bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type PayrollRunRecord struct {
	ID              int64
	OwnerID         int64
	BusinessID      int64
	PeriodStart     string
	PeriodEnd       string
	Status          string
	CreatedBy       int64
	ApprovedBy      int64
	ExportedAt      time.Time
	TotalGrossCents int64
	TotalNetCents   int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type PayrollEntryRecord struct {
	ID               int64
	OwnerID          int64
	BusinessID       int64
	PayrollRunID     int64
	EmployeeID       int64
	HoursWorked      float64
	OvertimeHours    float64
	BasePayCents     int64
	OvertimePayCents int64
	BonusesCents     int64
	DeductionsCents  int64
	NetPayCents      int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
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
