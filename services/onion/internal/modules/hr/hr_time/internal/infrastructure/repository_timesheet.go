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
	var employeeID sql.NullInt64
	if params.EmployeeID != nil {
		employeeID = sql.NullInt64{Int64: *params.EmployeeID, Valid: true}
	}
	var status sql.NullString
	if params.Status != nil {
		status = sql.NullString{String: *params.Status, Valid: true}
	}

	total, err := r.queries.CountTimesheets(ctx, sqlc.CountTimesheetsParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
		Status:     status,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to count timesheets", logger.Any("business_id", businessID), logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListTimesheets(ctx, sqlc.ListTimesheetsParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
		Status:     status,
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

// AdvanceApprovalLevel moves a submitted timesheet from expectedLevel to the
// next one and records the approval action in the same transaction. The
// UPDATE is conditional on status='submitted' AND
// current_approval_level = expectedLevel, where expectedLevel is the level the
// caller read before deciding to approve: an overlapping or retried approval
// then affects 0 rows and is reported as a conflict instead of incrementing
// twice and skipping a level of the chain. hr_leave and hr_overtime guard
// their advance queries the same way.
func (r *Repository) AdvanceApprovalLevel(ctx context.Context, businessID, timesheetID int64, expectedLevel int32, approverID int64) (*domain.TimesheetRecord, error) {
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

	result, err := qtx.AdvanceApprovalLevel(ctx, sqlc.AdvanceApprovalLevelParams{
		ID:                   timesheetID,
		BusinessID:           businessID,
		CurrentApprovalLevel: expectedLevel,
	})
	if err != nil {
		_ = tx.Rollback()
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

	if err := qtx.CreateApprovalAction(ctx, sqlc.CreateApprovalActionParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		TimesheetID: timesheetID,
		ApproverID:  approverID,
		Action:      "approved",
		Level:       expectedLevel,
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
