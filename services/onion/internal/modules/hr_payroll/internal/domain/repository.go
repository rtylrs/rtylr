package domain

import (
	"context"

	"github.com/voxire/onion/internal/platform/financejournal"
)

type Repository interface {
	// Payroll Config
	GetPayrollConfig(ctx context.Context, ownerID int64) (*PayrollConfigRecord, error)
	UpsertPayrollConfig(ctx context.Context, ownerID int64, cfg PayrollConfigRecord) error

	// Payroll Runs
	ListPayrollRuns(ctx context.Context, ownerID int64, params ListPayrollRunsParams) ([]PayrollRunRecord, int64, error)
	GetPayrollRun(ctx context.Context, ownerID, runID int64) (*PayrollRunRecord, error)
	GetPayrollRunByPeriod(ctx context.Context, ownerID int64, periodStart, periodEnd string) (*PayrollRunRecord, error)
	CreatePayrollRunWithEntries(ctx context.Context, ownerID int64, run PayrollRunRecord, entries []PayrollEntryRecord) (*PayrollRunRecord, []PayrollEntryRecord, error)
	UpdatePayrollRunStatus(ctx context.Context, ownerID, runID int64, status string) error
	SetPayrollRunApproved(ctx context.Context, ownerID, runID, approverID int64) error
	SetPayrollRunExported(ctx context.Context, ownerID, runID int64) error
	ClaimPayrollRunForExport(ctx context.Context, ownerID, runID int64) error
	ReleasePayrollRunExportClaim(ctx context.Context, ownerID, runID int64) error
	ValidatePayrollRunEmployees(ctx context.Context, ownerID, runID int64) error

	// Payroll Entries
	ListPayrollEntries(ctx context.Context, ownerID, runID int64) ([]PayrollEntryRecord, error)
	GetPayrollEntryForEmployee(ctx context.Context, ownerID, runID, employeeID int64) (*PayrollEntryRecord, error)

	// Self-service (my-runs)
	ListPayrollRunsForEmployee(ctx context.Context, ownerID, employeeID int64, params ListPayrollRunsParams) ([]PayrollRunRecord, int64, error)

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

// PermissionProvider lets the payroll service resolve which hr_employees row
// belongs to the calling platform user and check role-based permissions,
// without depending on hr_core directly.
type PermissionProvider interface {
	GetEmployeeIDByUserID(ctx context.Context, ownerID, userID int64) (int64, error)
	HasPermission(ctx context.Context, ownerID, employeeID int64, permissionKey string) (bool, error)
	HasRole(ctx context.Context, ownerID, employeeID, roleID int64) (bool, error)
}

// FinanceJournalPoster lets the payroll service hand a completed run off to the
// finance module without depending on it directly. The finance module is
// expected to translate the call into the appropriate journal entries
// (debit payroll expense, credit payable / cash) inside its own bounded
// context. Implementations must be idempotent on (ownerID, runID).
type FinanceJournalPoster = financejournal.PayrollJournalPoster

// PayrollJournalEntry is the payload handed to the finance module when a
// payroll run is exported. Money figures are passed in the currency stored
// on the run record; finance is responsible for any conversion.
type PayrollJournalEntry = financejournal.PayrollJournalEntry

// NotificationDispatcher lets the payroll service emit notifications without
// depending on the notification module directly. Implementations should be
// non-blocking from the caller's perspective (queue + return).
type NotificationDispatcher interface {
	Dispatch(ctx context.Context, event NotificationEvent) error
}

// NotificationEvent is the cross-module notification envelope. Type is a
// stable string key (e.g. "payroll.payslip", "payroll.period_open") that the
// notification module maps to a template + delivery channel based on company
// config.
type NotificationEvent struct {
	OwnerID int64
	Type    string
	Subject string
	Payload map[string]any
	UserIDs []int64 // optional explicit recipients; empty = use type-based routing
}
