package domain

import "context"

type Repository interface {
	CreateTimeEntry(ctx context.Context, ownerID int64, entry TimeEntryRecord) (*TimeEntryRecord, error)
	UpdateTimeEntry(ctx context.Context, ownerID int64, entry TimeEntryRecord) (*TimeEntryRecord, error)
	DeleteTimeEntry(ctx context.Context, ownerID, entryID int64) error
	GetTimeEntry(ctx context.Context, ownerID, entryID int64) (*TimeEntryRecord, error)
	ListTimeEntries(ctx context.Context, ownerID int64, params ListTimeEntriesParams) ([]TimeEntryRecord, int64, error)

	GetTimesheet(ctx context.Context, ownerID, timesheetID int64) (*TimesheetRecord, error)
	GetTimesheetForPeriod(ctx context.Context, ownerID, employeeID int64, periodStart string) (*TimesheetRecord, error)
	UpsertTimesheet(ctx context.Context, ownerID int64, ts TimesheetRecord) (*TimesheetRecord, error)
	ListTimesheets(ctx context.Context, ownerID int64, params ListTimesheetsParams) ([]TimesheetRecord, int64, error)

	SubmitTimesheet(ctx context.Context, ownerID, timesheetID int64) (*TimesheetRecord, error)
	AdvanceApprovalLevel(ctx context.Context, ownerID, timesheetID, approverID int64) (*TimesheetRecord, error)
	RejectTimesheet(ctx context.Context, ownerID, timesheetID int64, reason string) (*TimesheetRecord, error)
	LockTimesheet(ctx context.Context, ownerID, timesheetID int64) error

	CreateApprovalAction(ctx context.Context, ownerID int64, action ApprovalActionRecord) error

	ListUnsubmittedEmployees(ctx context.Context, ownerID int64, periodStart, periodEnd string) ([]int64, error)

	GetTimeEntryBySourceAndTask(ctx context.Context, ownerID, employeeID int64, source string, taskID int64) (*TimeEntryRecord, error)
	GetExistingTaskIDsBySource(ctx context.Context, ownerID, employeeID int64, source string, taskIDs []int64) (map[int64]struct{}, error)
	ListApprovedBillableEntries(ctx context.Context, ownerID int64) ([]TimeEntryRecord, error)
	MarkEntryExported(ctx context.Context, ownerID, entryID, billableInvoiceID int64) error
}

type ApprovalChainProvider interface {
	GetApprovalChain(ctx context.Context, ownerID int64, requestType string) (*ApprovalChainInfo, error)
}

type FlowTaskReader interface {
	GetCompletedTasksForEmployee(ctx context.Context, ownerID, employeeID int64, periodStart, periodEnd string) ([]FlowTaskRecord, error)
}
