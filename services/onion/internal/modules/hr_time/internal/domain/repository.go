package domain

import "context"

type Repository interface {
	MarkEntriesInvoiced(ctx context.Context, businessID, invoiceID int64, entryIDs []int64) (int64, error)
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
	AdvanceApprovalLevel(ctx context.Context, businessID, timesheetID int64, expectedLevel int32, approverID int64) (*TimesheetRecord, error)
	RejectTimesheet(ctx context.Context, businessID, timesheetID int64, reason string) (*TimesheetRecord, error)
	LockTimesheet(ctx context.Context, businessID, timesheetID int64) error
	LockTimesheetWithApprovalAction(ctx context.Context, businessID, timesheetID, approverID int64, level int32) (*TimesheetRecord, error)

	CreateApprovalAction(ctx context.Context, businessID int64, action ApprovalActionRecord) error

	ListTimesheetBusinessIDs(ctx context.Context) ([]int64, error)
	ListUnsubmittedEmployees(ctx context.Context, businessID int64, periodStart, periodEnd string) ([]int64, error)
}

type ApprovalChainProvider interface {
	GetApprovalChain(ctx context.Context, businessID int64, requestType string) (*ApprovalChainInfo, error)
}
