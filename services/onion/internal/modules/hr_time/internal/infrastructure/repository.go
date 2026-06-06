package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_time/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
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

func (r *Repository) ListTimeEntries(ctx context.Context, ownerID int64, params domain.ListTimeEntriesParams) ([]domain.TimeEntryRecord, int64, error) {
	periodStart, err := time.Parse("2006-01-02", params.PeriodStart)
	if err != nil {
		return nil, 0, err
	}
	periodEnd, err := time.Parse("2006-01-02", params.PeriodEnd)
	if err != nil {
		return nil, 0, err
	}

	total, err := r.queries.CountTimeEntries(ctx, sqlc.CountTimeEntriesParams{
		OwnerID:    ownerID,
		EmployeeID: params.EmployeeID,
		Date:       periodStart,
		Date_2:     periodEnd,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to count time entries", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListTimeEntries(ctx, sqlc.ListTimeEntriesParams{
		OwnerID:    ownerID,
		EmployeeID: params.EmployeeID,
		Date:       periodStart,
		Date_2:     periodEnd,
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list time entries", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, 0, err
	}

	entries := make([]domain.TimeEntryRecord, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, timeEntryFromRow(row))
	}
	return entries, total, nil
}

func (r *Repository) GetTimeEntry(ctx context.Context, ownerID, entryID int64) (*domain.TimeEntryRecord, error) {
	row, err := r.queries.GetTimeEntry(ctx, sqlc.GetTimeEntryParams{OwnerID: ownerID, ID: entryID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTimeEntryNotFound
		}
		logger.Ctx(ctx).Error("failed to get time entry", logger.Any("owner_id", ownerID), logger.Any("entry_id", entryID), logger.Error(err))
		return nil, err
	}
	entry := timeEntryFromRow(row)
	return &entry, nil
}

func (r *Repository) CreateTimeEntry(ctx context.Context, ownerID int64, entry domain.TimeEntryRecord) (*domain.TimeEntryRecord, error) {
	entryDate, err := time.Parse("2006-01-02", entry.Date)
	if err != nil {
		return nil, err
	}

	result, err := r.queries.CreateTimeEntry(ctx, sqlc.CreateTimeEntryParams{
		OwnerID:         ownerID,
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
		logger.Ctx(ctx).Error("failed to create time entry", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	entry.ID = id
	entry.OwnerID = ownerID
	return &entry, nil
}

func (r *Repository) UpdateTimeEntry(ctx context.Context, ownerID int64, entry domain.TimeEntryRecord) (*domain.TimeEntryRecord, error) {
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
		OwnerID:         ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to update time entry", logger.Any("owner_id", ownerID), logger.Any("entry_id", entry.ID), logger.Error(err))
		return nil, err
	}
	entry.OwnerID = ownerID
	return &entry, nil
}

func (r *Repository) DeleteTimeEntry(ctx context.Context, ownerID, entryID int64) error {
	return r.queries.DeleteTimeEntry(ctx, sqlc.DeleteTimeEntryParams{
		OwnerID: ownerID,
		ID:      entryID,
	})
}

// --- Timesheets ---

func (r *Repository) GetTimesheet(ctx context.Context, ownerID, timesheetID int64) (*domain.TimesheetRecord, error) {
	row, err := r.queries.GetTimesheet(ctx, sqlc.GetTimesheetParams{OwnerID: ownerID, ID: timesheetID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTimesheetNotFound
		}
		logger.Ctx(ctx).Error("failed to get timesheet", logger.Any("owner_id", ownerID), logger.Any("timesheet_id", timesheetID), logger.Error(err))
		return nil, err
	}
	ts := timesheetFromRow(row)
	return &ts, nil
}

func (r *Repository) GetTimesheetForPeriod(ctx context.Context, ownerID, employeeID int64, periodStart string) (*domain.TimesheetRecord, error) {
	ps, err := time.Parse("2006-01-02", periodStart)
	if err != nil {
		return nil, err
	}
	row, err := r.queries.GetTimesheetForPeriod(ctx, sqlc.GetTimesheetForPeriodParams{
		OwnerID:     ownerID,
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

func (r *Repository) UpsertTimesheet(ctx context.Context, ownerID int64, ts domain.TimesheetRecord) (*domain.TimesheetRecord, error) {
	ps, err := time.Parse("2006-01-02", ts.PeriodStart)
	if err != nil {
		return nil, err
	}
	pe, err := time.Parse("2006-01-02", ts.PeriodEnd)
	if err != nil {
		return nil, err
	}

	result, err := r.queries.CreateTimesheet(ctx, sqlc.CreateTimesheetParams{
		OwnerID:              ownerID,
		EmployeeID:           ts.EmployeeID,
		PeriodStart:          ps,
		PeriodEnd:            pe,
		Status:               ts.Status,
		CurrentApprovalLevel: ts.CurrentApprovalLevel,
		TotalMinutes:         ts.TotalMinutes,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create timesheet", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	ts.ID = id
	ts.OwnerID = ownerID
	return &ts, nil
}

func (r *Repository) ListTimesheets(ctx context.Context, ownerID int64, params domain.ListTimesheetsParams) ([]domain.TimesheetRecord, int64, error) {
	total, err := r.queries.CountTimesheets(ctx, sqlc.CountTimesheetsParams{
		OwnerID:    ownerID,
		EmployeeID: params.EmployeeID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to count timesheets", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListTimesheets(ctx, sqlc.ListTimesheetsParams{
		OwnerID:    ownerID,
		EmployeeID: params.EmployeeID,
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list timesheets", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, 0, err
	}

	timesheets := make([]domain.TimesheetRecord, 0, len(rows))
	for _, row := range rows {
		timesheets = append(timesheets, timesheetFromRow(row))
	}
	return timesheets, total, nil
}

func (r *Repository) SubmitTimesheet(ctx context.Context, ownerID, timesheetID int64) (*domain.TimesheetRecord, error) {
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

	if err := qtx.SubmitTimesheet(ctx, sqlc.SubmitTimesheetParams{ID: timesheetID, OwnerID: ownerID}); err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("failed to submit timesheet", logger.Any("owner_id", ownerID), logger.Any("timesheet_id", timesheetID), logger.Error(err))
		return nil, err
	}

	if err := qtx.CreateApprovalAction(ctx, sqlc.CreateApprovalActionParams{
		OwnerID:     ownerID,
		TimesheetID: timesheetID,
		ApproverID:  0,
		Action:      "submitted",
		Level:       0,
		Comment:     nullString(""),
	}); err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("failed to create submission approval action", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetTimesheet(ctx, ownerID, timesheetID)
}

func (r *Repository) AdvanceApprovalLevel(ctx context.Context, ownerID, timesheetID, approverID int64) (*domain.TimesheetRecord, error) {
	ts, err := r.GetTimesheet(ctx, ownerID, timesheetID)
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

	if err := qtx.AdvanceApprovalLevel(ctx, sqlc.AdvanceApprovalLevelParams{ID: timesheetID, OwnerID: ownerID}); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := qtx.CreateApprovalAction(ctx, sqlc.CreateApprovalActionParams{
		OwnerID:     ownerID,
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

	return r.GetTimesheet(ctx, ownerID, timesheetID)
}

func (r *Repository) RejectTimesheet(ctx context.Context, ownerID, timesheetID int64, reason string) (*domain.TimesheetRecord, error) {
	err := r.queries.RejectTimesheet(ctx, sqlc.RejectTimesheetParams{
		RejectionReason: nullString(reason),
		ID:              timesheetID,
		OwnerID:         ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to reject timesheet", logger.Any("owner_id", ownerID), logger.Any("timesheet_id", timesheetID), logger.Error(err))
		return nil, err
	}
	return r.GetTimesheet(ctx, ownerID, timesheetID)
}

func (r *Repository) LockTimesheet(ctx context.Context, ownerID, timesheetID int64) error {
	return r.queries.LockTimesheet(ctx, sqlc.LockTimesheetParams{
		ID:      timesheetID,
		OwnerID: ownerID,
	})
}

func (r *Repository) CreateApprovalAction(ctx context.Context, ownerID int64, action domain.ApprovalActionRecord) error {
	return r.queries.CreateApprovalAction(ctx, sqlc.CreateApprovalActionParams{
		OwnerID:     ownerID,
		TimesheetID: action.TimesheetID,
		ApproverID:  action.ApproverID,
		Action:      action.Action,
		Level:       action.Level,
		Comment:     nullString(action.Comment),
	})
}

func (r *Repository) GetTimeEntryBySourceAndTask(ctx context.Context, ownerID, employeeID int64, source string, taskID int64) (*domain.TimeEntryRecord, error) {
	row, err := r.queries.GetTimeEntryBySourceAndTask(ctx, sqlc.GetTimeEntryBySourceAndTaskParams{
		OwnerID:    ownerID,
		EmployeeID: employeeID,
		Source:     source,
		TaskID:     nullInt64(taskID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	entry := timeEntryFromRow(row)
	return &entry, nil
}

// GetExistingTaskIDsBySource returns the set of task IDs (from the given
// candidate list) that already have a time entry for the owner/employee/source
// combination. This batches what would otherwise be a per-task existence check
// into a single WHERE task_id IN (...) query.
func (r *Repository) GetExistingTaskIDsBySource(ctx context.Context, ownerID, employeeID int64, source string, taskIDs []int64) (map[int64]struct{}, error) {
	existing := make(map[int64]struct{}, len(taskIDs))
	if len(taskIDs) == 0 {
		return existing, nil
	}

	sqlStr, args, err := sq.Select("task_id").
		From("hr_time_entries").
		Where(sq.Eq{
			"owner_id":    ownerID,
			"employee_id": employeeID,
			"source":      source,
			"task_id":     taskIDs,
		}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("failed to build existing task entries query",
			logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		logger.Ctx(ctx).Error("failed to batch-check existing task entries",
			logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var taskID sql.NullInt64
		if scanErr := rows.Scan(&taskID); scanErr != nil {
			return nil, scanErr
		}
		if taskID.Valid {
			existing[taskID.Int64] = struct{}{}
		}
	}

	return existing, rows.Err()
}

func (r *Repository) ListApprovedBillableEntries(ctx context.Context, ownerID int64) ([]domain.TimeEntryRecord, error) {
	rows, err := r.queries.ListApprovedBillableEntries(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list approved billable entries", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, err
	}
	entries := make([]domain.TimeEntryRecord, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, timeEntryFromRow(row))
	}
	return entries, nil
}

func (r *Repository) MarkEntryExported(ctx context.Context, ownerID, entryID, billableInvoiceID int64) error {
	return r.queries.MarkEntryExported(ctx, sqlc.MarkEntryExportedParams{
		BillableInvoiceID: nullInt64(billableInvoiceID),
		ID:                entryID,
		OwnerID:           ownerID,
	})
}

func (r *Repository) ListUnsubmittedEmployees(ctx context.Context, ownerID int64, periodStart, periodEnd string) ([]int64, error) {
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
		OwnerID:     ownerID,
		Date:        ps,
		Date_2:      pe,
	})
}

// --- Helpers ---

func timeEntryFromRow(row sqlc.HrTimeEntry) domain.TimeEntryRecord {
	return domain.TimeEntryRecord{
		ID:                row.ID,
		OwnerID:           row.OwnerID,
		EmployeeID:        row.EmployeeID,
		TimesheetID:       row.TimesheetID.Int64,
		ProjectID:         row.ProjectID.Int64,
		TaskID:            row.TaskID.Int64,
		CategoryID:        row.CategoryID.Int64,
		Date:              row.Date.Format("2006-01-02"),
		DurationMinutes:   row.DurationMinutes,
		StartTime:         row.StartTime.String,
		EndTime:           row.EndTime.String,
		Description:       row.Description.String,
		IsBillable:        row.IsBillable,
		Status:            row.Status,
		RejectionReason:   row.RejectionReason.String,
		ApprovedBy:        row.ApprovedBy.Int64,
		Source:            row.Source,
		BillableInvoiceID: row.BillableInvoiceID.Int64,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}
}

func timesheetFromRow(row sqlc.HrTimesheet) domain.TimesheetRecord {
	return domain.TimesheetRecord{
		ID:                   row.ID,
		OwnerID:              row.OwnerID,
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

func nullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}

func nullString(v string) sql.NullString {
	return sql.NullString{String: v, Valid: v != ""}
}
