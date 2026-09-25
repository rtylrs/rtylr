package domain

import "time"

const (
	ReportTypeHecount           = "headcount"
	ReportTypeAttendanceSummary = "attendance_summary"
	ReportTypeLeaveSummary      = "leave_summary"
	ReportTypePayrollSummary    = "payroll_summary"
	ReportTypeOvertimeSummary   = "overtime_summary"
	ReportTypeTimesheetSummary  = "timesheet_summary"
)

var SupportedReportTypes = map[string]bool{
	ReportTypeHecount:           true,
	ReportTypeAttendanceSummary: true,
	ReportTypeLeaveSummary:      true,
	ReportTypePayrollSummary:    true,
	ReportTypeOvertimeSummary:   true,
	ReportTypeTimesheetSummary:  true,
}

type ReportParams struct {
	ActorID      int64
	OwnerID      int64
	BusinessID   int64
	DateFrom     time.Time
	DateTo       time.Time
	LocationID   *int64
	DepartmentID *int64
	EmployeeID   *int64
	Format       string
}
