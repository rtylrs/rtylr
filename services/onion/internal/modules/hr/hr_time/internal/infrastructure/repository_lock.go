package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

// LockTimesheetWithApprovalAction locks a submitted timesheet and records the
// final approval action in one transaction, the way SubmitTimesheet pairs its
// status change with its action. The service used to commit the lock first and
// only log a failed action insert, which left a locked timesheet whose approval
// history was permanently missing the final entry.
func (r *Repository) LockTimesheetWithApprovalAction(ctx context.Context, businessID, timesheetID, approverID int64, level int32) (*domain.TimesheetRecord, error) {
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

	result, err := qtx.LockTimesheet(ctx, sqlc.LockTimesheetParams{ID: timesheetID, BusinessID: businessID})
	if err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("failed to lock timesheet", logger.Any("business_id", businessID), logger.Any("timesheet_id", timesheetID), logger.Error(err))
		return nil, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if affected == 0 {
		_ = tx.Rollback()
		return nil, domain.ErrTimesheetNotPending
	}

	ts, err := qtx.GetTimesheet(ctx, sqlc.GetTimesheetParams{ID: timesheetID, BusinessID: businessID})
	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTimesheetNotFound
		}
		return nil, err
	}
	if _, err := qtx.ApproveTimesheetEntries(ctx, sqlc.ApproveTimesheetEntriesParams{
		TimesheetID:   nullInt64(timesheetID),
		ApprovedBy:    nullInt64(approverID),
		BusinessID:    businessID,
		EmployeeID:    ts.EmployeeID,
		TimesheetID_2: nullInt64(timesheetID),
		Date:          ts.PeriodStart,
		Date_2:        ts.PeriodEnd,
	}); err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("failed to approve timesheet entries", logger.Any("business_id", businessID), logger.Any("timesheet_id", timesheetID), logger.Error(err))
		return nil, err
	}

	if err := qtx.CreateApprovalAction(ctx, sqlc.CreateApprovalActionParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		TimesheetID: timesheetID,
		ApproverID:  approverID,
		Action:      "final_approved",
		Level:       level,
		Comment:     nullString(""),
	}); err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("failed to record final approval action", logger.Any("business_id", businessID), logger.Any("timesheet_id", timesheetID), logger.Error(err))
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetTimesheet(ctx, businessID, timesheetID)
}

func (r *Repository) RejectTimesheet(ctx context.Context, businessID, timesheetID int64, reason string) (*domain.TimesheetRecord, error) {
	affected, err := r.queries.RejectTimesheet(ctx, sqlc.RejectTimesheetParams{
		RejectionReason: nullString(reason),
		ID:              timesheetID,
		BusinessID:      businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to reject timesheet", logger.Any("business_id", businessID), logger.Any("timesheet_id", timesheetID), logger.Error(err))
		return nil, err
	}
	if affected == 0 {
		return nil, domain.ErrTimesheetNotPending
	}
	return r.GetTimesheet(ctx, businessID, timesheetID)
}

// LockTimesheet locks a submitted timesheet. It reports ErrTimesheetNotPending
// when the UPDATE matched nothing (draft, rejected, or already locked), so a
// caller cannot move a row the approval flow does not own.
func (r *Repository) LockTimesheet(ctx context.Context, businessID, timesheetID int64) error {
	result, err := r.queries.LockTimesheet(ctx, sqlc.LockTimesheetParams{
		ID:         timesheetID,
		BusinessID: businessID,
	})
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrTimesheetNotPending
	}
	return nil
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

func timeEntryFromBillableListRow(row sqlc.ListUnbilledBillableTimeEntriesRow) domain.TimeEntryRecord {
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
