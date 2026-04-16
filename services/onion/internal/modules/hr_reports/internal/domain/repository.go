package domain

import "context"

type Repository interface {
	GenerateHeadcountReport(ctx context.Context, ownerID int64, params ReportParams) ([]map[string]interface{}, error)
	GenerateAttendanceSummaryReport(ctx context.Context, ownerID int64, params ReportParams) ([]map[string]interface{}, error)
	GenerateLeaveSummaryReport(ctx context.Context, ownerID int64, params ReportParams) ([]map[string]interface{}, error)
	GeneratePayrollSummaryReport(ctx context.Context, ownerID int64, params ReportParams) ([]map[string]interface{}, error)
	GenerateOvertimeSummaryReport(ctx context.Context, ownerID int64, params ReportParams) ([]map[string]interface{}, error)
	GenerateTimesheetSummaryReport(ctx context.Context, ownerID int64, params ReportParams) ([]map[string]interface{}, error)
}
