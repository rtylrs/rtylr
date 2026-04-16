package domain

import "context"

type Repository interface {
	// Payroll Config
	GetPayrollConfig(ctx context.Context, ownerID int64) (*PayrollConfigRecord, error)
	UpsertPayrollConfig(ctx context.Context, ownerID int64, cfg PayrollConfigRecord) error

	// Payroll Runs
	ListPayrollRuns(ctx context.Context, ownerID int64, params ListPayrollRunsParams) ([]PayrollRunRecord, int64, error)
	GetPayrollRun(ctx context.Context, ownerID, runID int64) (*PayrollRunRecord, error)
	CreatePayrollRunWithEntries(ctx context.Context, ownerID int64, run PayrollRunRecord, entries []PayrollEntryRecord) (*PayrollRunRecord, []PayrollEntryRecord, error)
	UpdatePayrollRunStatus(ctx context.Context, ownerID, runID int64, status string) error
	SetPayrollRunApproved(ctx context.Context, ownerID, runID, approverID int64) error
	SetPayrollRunExported(ctx context.Context, ownerID, runID int64) error

	// Payroll Entries
	ListPayrollEntries(ctx context.Context, ownerID, runID int64) ([]PayrollEntryRecord, error)

	// Cross-module data gathering
	ListApprovedTimesheetSummaries(ctx context.Context, ownerID int64, periodStart, periodEnd string) ([]TimesheetSummary, error)
	ListApprovedOvertimeSummaries(ctx context.Context, ownerID int64, periodStart, periodEnd string) ([]OvertimeSummary, error)
	ListEmployeePayInfo(ctx context.Context, ownerID int64) ([]EmployeePayInfo, error)
	GetOvertimeConfig(ctx context.Context, ownerID int64) (float64, bool, error)

	// Background job helpers
	ListDistinctOwnerIDs(ctx context.Context) ([]int64, error)
}

type ApprovalChainProvider interface {
	GetApprovalChain(ctx context.Context, ownerID int64, requestType string) (*ApprovalChainInfo, error)
}
