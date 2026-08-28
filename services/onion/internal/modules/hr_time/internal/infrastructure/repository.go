package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_time/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

// --- Time Entries ---

// nullDate parses an optional ISO-8601 day into the nullable form sqlc expects.
// A malformed value is an error rather than a silently dropped filter: a caller
// who mistypes a date should not receive the unfiltered set.
func nullDate(value *string) (sql.NullTime, error) {
	if value == nil || *value == "" {
		return sql.NullTime{}, nil
	}
	parsed, err := time.Parse("2006-01-02", *value)
	if err != nil {
		return sql.NullTime{}, err
	}
	return sql.NullTime{Time: parsed, Valid: true}, nil
}

func (r *Repository) ListTimeEntries(ctx context.Context, businessID int64, params domain.ListTimeEntriesParams) ([]domain.TimeEntryRecord, int64, error) {
	fromDate, err := nullDate(params.FromDate)
	if err != nil {
		return nil, 0, err
	}
	toDate, err := nullDate(params.ToDate)
	if err != nil {
		return nil, 0, err
	}

	filters := struct {
		employeeID sql.NullInt64
		projectID  sql.NullInt64
		isBillable sql.NullBool
		status     sql.NullString
	}{}
	if params.EmployeeID != nil {
		filters.employeeID = sql.NullInt64{Int64: *params.EmployeeID, Valid: true}
	}
	if params.ProjectID != nil {
		filters.projectID = sql.NullInt64{Int64: *params.ProjectID, Valid: true}
	}
	if params.IsBillable != nil {
		filters.isBillable = sql.NullBool{Bool: *params.IsBillable, Valid: true}
	}
	if params.Status != nil {
		filters.status = sql.NullString{String: *params.Status, Valid: true}
	}

	total, err := r.queries.CountTimeEntries(ctx, sqlc.CountTimeEntriesParams{
		BusinessID: businessID,
		EmployeeID: filters.employeeID,
		FromDate:   fromDate,
		ToDate:     toDate,
		ProjectID:  filters.projectID,
		IsBillable: filters.isBillable,
		Status:     filters.status,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to count time entries", logger.Any("business_id", businessID), logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListTimeEntries(ctx, sqlc.ListTimeEntriesParams{
		BusinessID: businessID,
		EmployeeID: filters.employeeID,
		FromDate:   fromDate,
		ToDate:     toDate,
		ProjectID:  filters.projectID,
		IsBillable: filters.isBillable,
		Status:     filters.status,
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list time entries", logger.Any("business_id", businessID), logger.Error(err))
		return nil, 0, err
	}

	entries := make([]domain.TimeEntryRecord, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, timeEntryFromListRow(row))
	}
	return entries, total, nil
}

func (r *Repository) GetTimeEntry(ctx context.Context, businessID, entryID int64) (*domain.TimeEntryRecord, error) {
	row, err := r.queries.GetTimeEntry(ctx, sqlc.GetTimeEntryParams{BusinessID: businessID, ID: entryID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTimeEntryNotFound
		}
		logger.Ctx(ctx).Error("failed to get time entry", logger.Any("business_id", businessID), logger.Any("entry_id", entryID), logger.Error(err))
		return nil, err
	}
	entry := timeEntryFromGetRow(row)
	return &entry, nil
}

func (r *Repository) CreateTimeEntry(ctx context.Context, businessID int64, entry domain.TimeEntryRecord) (*domain.TimeEntryRecord, error) {
	entryDate, err := time.Parse("2006-01-02", entry.Date)
	if err != nil {
		return nil, err
	}

	result, err := r.queries.CreateTimeEntry(ctx, sqlc.CreateTimeEntryParams{
		OwnerID:         businessID,
		BusinessID:      businessID,
		EmployeeID:      entry.EmployeeID,
		TimesheetID:     nullInt64(entry.TimesheetID),
		ProjectID:       nullInt64(entry.ProjectID),
		TaskID:          nullInt64(entry.TaskID),
		CategoryID:      nullInt64(entry.CategoryID),
		Date:            entryDate,
		DurationMinutes: entry.DurationMinutes,
		StartTime:       nullString(entry.StartTime),
		EndTime:         nullString(entry.EndTime),
		Description:     nullString(entry.Description),
		IsBillable:      entry.IsBillable,
		Status:          entry.Status,
		RejectionReason: nullString(entry.RejectionReason),
		ApprovedBy:      nullInt64(entry.ApprovedBy),
		Source:          entry.Source,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create time entry", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	entry.ID = id
	entry.OwnerID = businessID
	entry.BusinessID = businessID
	return &entry, nil
}

func (r *Repository) UpdateTimeEntry(ctx context.Context, businessID int64, entry domain.TimeEntryRecord) (*domain.TimeEntryRecord, error) {
	entryDate, err := time.Parse("2006-01-02", entry.Date)
	if err != nil {
		return nil, err
	}

	err = r.queries.UpdateTimeEntry(ctx, sqlc.UpdateTimeEntryParams{
		EmployeeID:      entry.EmployeeID,
		TimesheetID:     nullInt64(entry.TimesheetID),
		ProjectID:       nullInt64(entry.ProjectID),
		TaskID:          nullInt64(entry.TaskID),
		CategoryID:      nullInt64(entry.CategoryID),
		Date:            entryDate,
		DurationMinutes: entry.DurationMinutes,
		StartTime:       nullString(entry.StartTime),
		EndTime:         nullString(entry.EndTime),
		Description:     nullString(entry.Description),
		IsBillable:      entry.IsBillable,
		Status:          entry.Status,
		RejectionReason: nullString(entry.RejectionReason),
		ApprovedBy:      nullInt64(entry.ApprovedBy),
		Source:          entry.Source,
		ID:              entry.ID,
		BusinessID:      businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to update time entry", logger.Any("business_id", businessID), logger.Any("entry_id", entry.ID), logger.Error(err))
		return nil, err
	}
	entry.BusinessID = businessID
	return &entry, nil
}

func (r *Repository) DeleteTimeEntry(ctx context.Context, businessID, entryID int64) error {
	return r.queries.DeleteTimeEntry(ctx, sqlc.DeleteTimeEntryParams{
		BusinessID: businessID,
		ID:         entryID,
	})
}

// --- Timesheets ---

func (r *Repository) GetTimesheet(ctx context.Context, businessID, timesheetID int64) (*domain.TimesheetRecord, error) {
	row, err := r.queries.GetTimesheet(ctx, sqlc.GetTimesheetParams{BusinessID: businessID, ID: timesheetID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTimesheetNotFound
		}
		logger.Ctx(ctx).Error("failed to get timesheet", logger.Any("business_id", businessID), logger.Any("timesheet_id", timesheetID), logger.Error(err))
		return nil, err
	}
	ts := timesheetFromRow(row)
	return &ts, nil
}

func (r *Repository) GetTimesheetForPeriod(ctx context.Context, businessID, employeeID int64, periodStart string) (*domain.TimesheetRecord, error) {
	ps, err := time.Parse("2006-01-02", periodStart)
	if err != nil {
		return nil, err
	}
	row, err := r.queries.GetTimesheetForPeriod(ctx, sqlc.GetTimesheetForPeriodParams{
		BusinessID:  businessID,
		EmployeeID:  employeeID,
		PeriodStart: ps,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	ts := timesheetFromRow(row)
	return &ts, nil
}

func (r *Repository) UpsertTimesheet(ctx context.Context, businessID int64, ts domain.TimesheetRecord) (*domain.TimesheetRecord, error) {
	ps, err := time.Parse("2006-01-02", ts.PeriodStart)
	if err != nil {
		return nil, err
	}
	pe, err := time.Parse("2006-01-02", ts.PeriodEnd)
	if err != nil {
		return nil, err
	}

	result, err := r.queries.CreateTimesheet(ctx, sqlc.CreateTimesheetParams{
		OwnerID:              businessID,
		BusinessID:           businessID,
		EmployeeID:           ts.EmployeeID,
		PeriodStart:          ps,
		PeriodEnd:            pe,
		Status:               ts.Status,
		CurrentApprovalLevel: ts.CurrentApprovalLevel,
		TotalMinutes:         ts.TotalMinutes,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create timesheet", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	ts.ID = id
	ts.OwnerID = businessID
	ts.BusinessID = businessID
	return &ts, nil
}

func (r *Repository) ListTimesheets(ctx context.Context, businessID int64, params domain.ListTimesheetsParams) ([]domain.TimesheetRecord, int64, error) {
	total, err := r.queries.CountTimesheets(ctx, sqlc.CountTimesheetsParams{
		BusinessID: businessID,
		EmployeeID: params.EmployeeID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to count timesheets", logger.Any("business_id", businessID), logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListTimesheets(ctx, sqlc.ListTimesheetsParams{
		BusinessID: businessID,
		EmployeeID: params.EmployeeID,
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list timesheets", logger.Any("business_id", businessID), logger.Error(err))
		return nil, 0, err
	}

	timesheets := make([]domain.TimesheetRecord, 0, len(rows))
	for _, row := range rows {
		timesheets = append(timesheets, timesheetFromRow(row))
	}
	return timesheets, total, nil
}

func (r *Repository) SubmitTimesheet(ctx context.Context, businessID, timesheetID int64) (*domain.TimesheetRecord, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	qtx := r.queries.WithTx(tx)

	if err := qtx.SubmitTimesheet(ctx, sqlc.SubmitTimesheetParams{ID: timesheetID, BusinessID: businessID}); err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("failed to submit timesheet", logger.Any("business_id", businessID), logger.Any("timesheet_id", timesheetID), logger.Error(err))
		return nil, err
	}

	if err := qtx.CreateApprovalAction(ctx, sqlc.CreateApprovalActionParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		TimesheetID: timesheetID,
		ApproverID:  0,
		Action:      "submitted",
		Level:       0,
		Comment:     nullString(""),
	}); err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("failed to create submission approval action", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetTimesheet(ctx, businessID, timesheetID)
}

func (r *Repository) AdvanceApprovalLevel(ctx context.Context, businessID, timesheetID, approverID int64) (*domain.TimesheetRecord, error) {
	ts, err := r.GetTimesheet(ctx, businessID, timesheetID)
	if err != nil {
		return nil, err
	}

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	qtx := r.queries.WithTx(tx)

	if err := qtx.AdvanceApprovalLevel(ctx, sqlc.AdvanceApprovalLevelParams{ID: timesheetID, BusinessID: businessID}); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := qtx.CreateApprovalAction(ctx, sqlc.CreateApprovalActionParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		TimesheetID: timesheetID,
		ApproverID:  approverID,
		Action:      "approved",
		Level:       ts.CurrentApprovalLevel,
		Comment:     nullString(""),
	}); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetTimesheet(ctx, businessID, timesheetID)
}

func (r *Repository) RejectTimesheet(ctx context.Context, businessID, timesheetID int64, reason string) (*domain.TimesheetRecord, error) {
	err := r.queries.RejectTimesheet(ctx, sqlc.RejectTimesheetParams{
		RejectionReason: nullString(reason),
		ID:              timesheetID,
		BusinessID:      businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to reject timesheet", logger.Any("business_id", businessID), logger.Any("timesheet_id", timesheetID), logger.Error(err))
		return nil, err
	}
	return r.GetTimesheet(ctx, businessID, timesheetID)
}

func (r *Repository) LockTimesheet(ctx context.Context, businessID, timesheetID int64) error {
	return r.queries.LockTimesheet(ctx, sqlc.LockTimesheetParams{
		ID:         timesheetID,
		BusinessID: businessID,
	})
}

func (r *Repository) CreateApprovalAction(ctx context.Context, businessID int64, action domain.ApprovalActionRecord) error {
	return r.queries.CreateApprovalAction(ctx, sqlc.CreateApprovalActionParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		TimesheetID: action.TimesheetID,
		ApproverID:  action.ApproverID,
		Action:      action.Action,
		Level:       action.Level,
		Comment:     nullString(action.Comment),
	})
}

func (r *Repository) ListTimesheetBusinessIDs(ctx context.Context) ([]int64, error) {
	rows, err := r.queries.ListTimesheetBusinessIDs(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list timesheet business IDs", logger.Error(err))
		return nil, err
	}

	businessIDs := make([]int64, 0, len(rows))
	businessIDs = append(businessIDs, rows...)
	return businessIDs, nil
}

func (r *Repository) ListUnsubmittedEmployees(ctx context.Context, businessID int64, periodStart, periodEnd string) ([]int64, error) {
	ps, err := time.Parse("2006-01-02", periodStart)
	if err != nil {
		return nil, err
	}
	pe, err := time.Parse("2006-01-02", periodEnd)
	if err != nil {
		return nil, err
	}
	return r.queries.ListUnsubmittedEmployees(ctx, sqlc.ListUnsubmittedEmployeesParams{
		PeriodStart: ps,
		BusinessID:  businessID,
		Date:        ps,
		Date_2:      pe,
	})
}

// --- Helpers ---

func timeEntryFromFields(
	id, ownerID, businessID, employeeID int64,
	timesheetID, projectID, taskID, categoryID sql.NullInt64,
	date time.Time,
	durationMinutes int32,
	startTime, endTime, description sql.NullString,
	isBillable bool,
	status string,
	rejectionReason sql.NullString,
	approvedBy sql.NullInt64,
	source string,
	billableInvoiceID sql.NullInt64,
	createdAt, updatedAt time.Time,
) domain.TimeEntryRecord {
	return domain.TimeEntryRecord{
		ID:                id,
		OwnerID:           ownerID,
		BusinessID:        businessID,
		EmployeeID:        employeeID,
		TimesheetID:       timesheetID.Int64,
		ProjectID:         projectID.Int64,
		TaskID:            taskID.Int64,
		CategoryID:        categoryID.Int64,
		Date:              date.Format("2006-01-02"),
		DurationMinutes:   durationMinutes,
		StartTime:         startTime.String,
		EndTime:           endTime.String,
		Description:       description.String,
		IsBillable:        isBillable,
		Status:            status,
		RejectionReason:   rejectionReason.String,
		ApprovedBy:        approvedBy.Int64,
		Source:            source,
		BillableInvoiceID: billableInvoiceID.Int64,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}
}

func timeEntryFromListRow(row sqlc.ListTimeEntriesRow) domain.TimeEntryRecord {
	return timeEntryFromFields(row.ID, row.OwnerID, row.BusinessID, row.EmployeeID, row.TimesheetID, row.ProjectID, row.TaskID, row.CategoryID, row.Date, row.DurationMinutes, row.StartTime, row.EndTime, row.Description, row.IsBillable, row.Status, row.RejectionReason, row.ApprovedBy, row.Source, row.BillableInvoiceID, row.CreatedAt, row.UpdatedAt)
}

func timeEntryFromGetRow(row sqlc.GetTimeEntryRow) domain.TimeEntryRecord {
	return timeEntryFromFields(row.ID, row.OwnerID, row.BusinessID, row.EmployeeID, row.TimesheetID, row.ProjectID, row.TaskID, row.CategoryID, row.Date, row.DurationMinutes, row.StartTime, row.EndTime, row.Description, row.IsBillable, row.Status, row.RejectionReason, row.ApprovedBy, row.Source, row.BillableInvoiceID, row.CreatedAt, row.UpdatedAt)
}

func timesheetFromRow(row sqlc.HrTimesheet) domain.TimesheetRecord {
	return domain.TimesheetRecord{
		ID:                   row.ID,
		OwnerID:              row.OwnerID,
		BusinessID:           row.BusinessID,
		EmployeeID:           row.EmployeeID,
		PeriodStart:          row.PeriodStart.Format("2006-01-02"),
		PeriodEnd:            row.PeriodEnd.Format("2006-01-02"),
		Status:               row.Status,
		CurrentApprovalLevel: row.CurrentApprovalLevel,
		SubmittedAt:          row.SubmittedAt.Time,
		TotalMinutes:         row.TotalMinutes,
		RejectionReason:      row.RejectionReason.String,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}
}

var nullInt64 = sqlutil.NullInt64

var nullString = sqlutil.NullString

// MarkEntriesInvoiced stamps each listed entry with the settling invoice id
// inside ONE transaction: the whole batch lands or none of it does, so a
// retry after any failure cannot double-mark. Rows affected are summed so the
// caller can report skipped ids (foreign, ineligible, or already invoiced do
// not match the UPDATE).
func (r *Repository) MarkEntriesInvoiced(ctx context.Context, businessID, invoiceID int64, entryIDs []int64) (marked int64, err error) {
	if len(entryIDs) == 0 {
		return 0, nil
	}
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	committed := false
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = tx.Rollback()
			panic(recovered)
		}
		if !committed {
			_ = tx.Rollback()
		}
	}()

	qtx := r.queries.WithTx(tx)
	for _, entryID := range entryIDs {
		n, execErr := qtx.MarkEntryInvoiced(ctx, sqlc.MarkEntryInvoicedParams{
			BillableInvoiceID: sql.NullInt64{Int64: invoiceID, Valid: true},
			BusinessID:        businessID,
			ID:                entryID,
		})
		if execErr != nil {
			logger.Ctx(ctx).Error("failed marking time entry invoiced",
				logger.Any("business_id", businessID),
				logger.Any("invoice_id", invoiceID),
				logger.Any("entry_id", entryID), logger.Error(execErr))
			return 0, execErr
		}
		marked += n
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	committed = true
	return marked, nil
}
