package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) ListUnbilledBillableTimeEntries(ctx context.Context, businessID int64, limit, offset int32) ([]domain.TimeEntryRecord, int64, int64, error) {
	totals, err := r.queries.CountUnbilledBillableTimeEntries(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to count unbilled billable time entries", logger.Any("business_id", businessID), logger.Error(err))
		return nil, 0, 0, err
	}

	rows, err := r.queries.ListUnbilledBillableTimeEntries(ctx, sqlc.ListUnbilledBillableTimeEntriesParams{
		BusinessID: businessID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list unbilled billable time entries", logger.Any("business_id", businessID), logger.Error(err))
		return nil, 0, 0, err
	}

	entries := make([]domain.TimeEntryRecord, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, timeEntryFromBillableListRow(row))
	}
	return entries, totals.Total, totals.TotalMinutes, nil
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
