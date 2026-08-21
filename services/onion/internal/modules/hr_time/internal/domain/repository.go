package domain

import "context"

type Repository interface {
	CreateTimeEntry(ctx context.Context, businessID int64, entry TimeEntryRecord) (*TimeEntryRecord, error)
	UpdateTimeEntry(ctx context.Context, businessID int64, entry TimeEntryRecord) (*TimeEntryRecord, error)
	DeleteTimeEntry(ctx context.Context, businessID, entryID int64) error
	GetTimeEntry(ctx context.Context, businessID, entryID int64) (*TimeEntryRecord, error)
	ListTimeEntries(ctx context.Context, businessID int64, params ListTimeEntriesParams) ([]TimeEntryRecord, int64, error)

	GetTimesheet(ctx context.Context, businessID, timesheetID int64) (*TimesheetRecord, error)
	GetTimesheetForPeriod(ctx context.Context, businessID, employeeID int64, periodStart string) (*TimesheetRecord, error)
	UpsertTimesheet(ctx context.Context, businessID int64, ts TimesheetRecord) (*TimesheetRecord, error)
	ListTimesheets(ctx context.Context, businessID int64, params ListTimesheetsParams) ([]TimesheetRecord, int64, error)

	SubmitTimesheet(ctx context.Context, businessID, timesheetID int64) (*TimesheetRecord, error)
	AdvanceApprovalLevel(ctx context.Context, businessID, timesheetID, approverID int64) (*TimesheetRecord, error)
	RejectTimesheet(ctx context.Context, businessID, timesheetID int64, reason string) (*TimesheetRecord, error)
	LockTimesheet(ctx context.Context, businessID, timesheetID int64) error

	CreateApprovalAction(ctx context.Context, businessID int64, action ApprovalActionRecord) error

	ListTimesheetBusinessIDs(ctx context.Context) ([]int64, error)
	ListUnsubmittedEmployees(ctx context.Context, businessID int64, periodStart, periodEnd string) ([]int64, error)

	ListApprovedBillableEntries(ctx context.Context, businessID int64) ([]TimeEntryRecord, error)
	MarkEntryExported(ctx context.Context, businessID, entryID, exportBatchID int64) error
	ExportApprovedBillableEntries(ctx context.Context, businessID, exportBatchID int64) ([]TimeEntryRecord, error)
}

type ApprovalChainProvider interface {
	GetApprovalChain(ctx context.Context, businessID int64, requestType string) (*ApprovalChainInfo, error)
}
