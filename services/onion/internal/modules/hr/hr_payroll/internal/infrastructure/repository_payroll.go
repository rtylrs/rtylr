package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

func (r *Repository) SetPayrollRunApproved(ctx context.Context, businessID, runID, approverID int64) error {
	result, err := r.queries.SetPayrollRunApproved(ctx, sqlc.SetPayrollRunApprovedParams{
		ApprovedBy: sql.NullInt64{Int64: approverID, Valid: true},
		BusinessID: businessID,
		ID:         runID,
	})
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return domain.ErrPayrollRunNotApprovable
	}
	return nil
}

func (r *Repository) SetPayrollRunExported(ctx context.Context, businessID, runID int64) error {
	result, err := r.queries.SetPayrollRunExported(ctx, sqlc.SetPayrollRunExportedParams{
		BusinessID: businessID,
		ID:         runID,
	})
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return domain.ErrPayrollRunNotApproved
	}
	return nil
}

func (r *Repository) ClaimPayrollRunForExport(ctx context.Context, businessID, runID int64) error {
	result, err := r.queries.ClaimPayrollRunForExport(ctx, sqlc.ClaimPayrollRunForExportParams{
		BusinessID: businessID,
		ID:         runID,
	})
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 1 {
		return nil
	}

	run, err := r.GetPayrollRun(ctx, businessID, runID)
	if err != nil {
		return err
	}
	if run.Status == domain.PayrollStatusExported || !run.ExportedAt.IsZero() {
		return domain.ErrPayrollAlreadyExported
	}
	if run.Status == domain.PayrollStatusExporting {
		return domain.ErrPayrollExportInProgress
	}
	return domain.ErrPayrollRunNotApproved
}

func (r *Repository) ReleasePayrollRunExportClaim(ctx context.Context, businessID, runID int64) error {
	return r.queries.ReleasePayrollRunExportClaim(ctx, sqlc.ReleasePayrollRunExportClaimParams{
		BusinessID: businessID,
		ID:         runID,
	})
}

func (r *Repository) ValidatePayrollRunEmployees(ctx context.Context, businessID, runID int64) error {
	inactiveCount, err := r.queries.CountInactivePayrollEmployees(ctx, sqlc.CountInactivePayrollEmployeesParams{
		BusinessID:   businessID,
		PayrollRunID: runID,
	})
	if err != nil {
		return err
	}
	if inactiveCount > 0 {
		return domain.ErrPayrollEmployeeInactive
	}
	return nil
}

// --- Payroll Entries ---

func (r *Repository) ListPayrollEntries(ctx context.Context, businessID, runID int64) ([]domain.PayrollEntryRecord, error) {
	rows, err := r.queries.ListPayrollEntries(ctx, sqlc.ListPayrollEntriesParams{
		BusinessID:   businessID,
		PayrollRunID: runID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list payroll entries", logger.Any("business_id", businessID), logger.Any("run_id", runID), logger.Error(err))
		return nil, err
	}

	entries := make([]domain.PayrollEntryRecord, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, payrollEntryFromRow(row))
	}
	return entries, nil
}

// ListPayrollRunsForEmployee lists payroll runs in which the given employee has
// an entry, restricted to statuses that represent finalized pay (approved or
// exported) — draft/in-review runs are not visible to self-service callers.
func (r *Repository) ListPayrollRunsForEmployee(ctx context.Context, businessID, employeeID int64, params domain.ListPayrollRunsParams) ([]domain.PayrollRunRecord, int64, error) {
	total, err := r.queries.CountPayrollRunsForEmployee(ctx, sqlc.CountPayrollRunsForEmployeeParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
		Status:     domain.PayrollStatusApproved,
		Status_2:   domain.PayrollStatusExported,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to count payroll runs for employee", logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListPayrollRunsForEmployee(ctx, sqlc.ListPayrollRunsForEmployeeParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
		Status:     domain.PayrollStatusApproved,
		Status_2:   domain.PayrollStatusExported,
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list payroll runs for employee", logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
		return nil, 0, err
	}

	runs := make([]domain.PayrollRunRecord, 0, len(rows))
	for _, row := range rows {
		runs = append(runs, payrollRunFromRow(row))
	}
	return runs, total, nil
}

// GetPayrollEntryForEmployee returns only the given employee's own entry from a
// payroll run, never another employee's. A missing entry (wrong employee, or
// the run doesn't exist) is reported as ErrPayrollRunNotFound so a client
// probing run IDs can't distinguish "no such run" from "not your entry".
func (r *Repository) GetPayrollEntryForEmployee(ctx context.Context, businessID, runID, employeeID int64) (*domain.PayrollEntryRecord, error) {
	row, err := r.queries.GetPayrollEntryForEmployee(ctx, sqlc.GetPayrollEntryForEmployeeParams{
		BusinessID:   businessID,
		PayrollRunID: runID,
		EmployeeID:   employeeID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPayrollRunNotFound
		}
		logger.Ctx(ctx).Error("failed to get payroll entry for employee", logger.Any("business_id", businessID), logger.Any("run_id", runID), logger.Any("employee_id", employeeID), logger.Error(err))
		return nil, err
	}
	entry := payrollEntryFromRow(row)
	return &entry, nil
}

// cross-schema read from hr_time — approved timesheet hours for payroll computation.
func (r *Repository) ListApprovedTimesheetSummaries(ctx context.Context, ownerID int64, periodStart, periodEnd string) ([]domain.TimesheetSummary, error) {
	ps, err := time.Parse("2006-01-02", periodStart)
	if err != nil {
		return nil, err
	}
	pe, err := time.Parse("2006-01-02", periodEnd)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListApprovedTimesheetSummaries(ctx, sqlc.ListApprovedTimesheetSummariesParams{
		BusinessID:  ownerID,
		PeriodStart: ps,
		PeriodEnd:   pe,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list timesheet summaries", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, err
	}

	summaries := make([]domain.TimesheetSummary, 0, len(rows))
	for _, row := range rows {
		summaries = append(summaries, domain.TimesheetSummary{
			EmployeeID:   row.EmployeeID,
			TotalMinutes: interfaceToInt32(row.TotalMinutes),
		})
	}
	return summaries, nil
}
