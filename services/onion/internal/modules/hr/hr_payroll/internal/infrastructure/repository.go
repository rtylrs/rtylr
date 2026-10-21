package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

const mysqlErrDuplicateEntry = 1062

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

// --- Payroll Config ---

func (r *Repository) GetPayrollConfig(ctx context.Context, businessID int64) (*domain.PayrollConfigRecord, error) {
	row, err := r.queries.GetPayrollConfig(ctx, businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPayrollNotConfigured
		}
		logger.Ctx(ctx).Error("failed to get payroll config", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	cfg := payrollConfigFromRow(row)
	return &cfg, nil
}

func (r *Repository) UpsertPayrollConfig(ctx context.Context, businessID int64, cfg domain.PayrollConfigRecord) error {
	_, err := r.queries.UpsertPayrollConfig(ctx, sqlc.UpsertPayrollConfigParams{
		OwnerID:         businessID,
		BusinessID:      businessID,
		PeriodType:      cfg.PeriodType,
		PayDayOfMonth:   cfg.PayDayOfMonth,
		OvertimeEnabled: cfg.OvertimeEnabled,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to upsert payroll config", logger.Any("business_id", businessID), logger.Error(err))
	}
	return err
}

// --- Payroll Runs ---

func (r *Repository) ListPayrollRuns(ctx context.Context, businessID int64, params domain.ListPayrollRunsParams) ([]domain.PayrollRunRecord, int64, error) {
	total, err := r.queries.CountPayrollRuns(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to count payroll runs", logger.Any("business_id", businessID), logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListPayrollRuns(ctx, sqlc.ListPayrollRunsParams{
		BusinessID: businessID,
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list payroll runs", logger.Any("business_id", businessID), logger.Error(err))
		return nil, 0, err
	}

	runs := make([]domain.PayrollRunRecord, 0, len(rows))
	for _, row := range rows {
		runs = append(runs, payrollRunFromRow(row))
	}
	return runs, total, nil
}

func (r *Repository) GetPayrollRun(ctx context.Context, businessID, runID int64) (*domain.PayrollRunRecord, error) {
	row, err := r.queries.GetPayrollRun(ctx, sqlc.GetPayrollRunParams{BusinessID: businessID, ID: runID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPayrollRunNotFound
		}
		logger.Ctx(ctx).Error("failed to get payroll run", logger.Any("business_id", businessID), logger.Any("run_id", runID), logger.Error(err))
		return nil, err
	}
	run := payrollRunFromRow(row)
	return &run, nil
}

func (r *Repository) GetPayrollRunByPeriod(ctx context.Context, businessID int64, periodStart, periodEnd string) (*domain.PayrollRunRecord, error) {
	ps, err := time.Parse("2006-01-02", periodStart)
	if err != nil {
		return nil, err
	}
	pe, err := time.Parse("2006-01-02", periodEnd)
	if err != nil {
		return nil, err
	}
	row, err := r.queries.GetPayrollRunByPeriod(ctx, sqlc.GetPayrollRunByPeriodParams{
		BusinessID:  businessID,
		PeriodStart: ps,
		PeriodEnd:   pe,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPayrollRunNotFound
		}
		logger.Ctx(ctx).Error("failed to get payroll run by period", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	run := payrollRunFromRow(row)
	return &run, nil
}

func (r *Repository) CreatePayrollRunWithEntries(ctx context.Context, businessID int64, run domain.PayrollRunRecord, entries []domain.PayrollEntryRecord) (*domain.PayrollRunRecord, []domain.PayrollEntryRecord, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	qtx := r.queries.WithTx(tx)

	periodStart, parseErr := time.Parse("2006-01-02", run.PeriodStart)
	if parseErr != nil {
		_ = tx.Rollback()
		return nil, nil, parseErr
	}
	periodEnd, parseErr := time.Parse("2006-01-02", run.PeriodEnd)
	if parseErr != nil {
		_ = tx.Rollback()
		return nil, nil, parseErr
	}

	result, err := qtx.CreatePayrollRun(ctx, sqlc.CreatePayrollRunParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Status:      run.Status,
		CreatedBy:   run.CreatedBy,
		TotalGross:  centsToDecimalString(run.TotalGrossCents),
		TotalNet:    centsToDecimalString(run.TotalNetCents),
	})
	if err != nil {
		_ = tx.Rollback()
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlErrDuplicateEntry {
			return nil, nil, domain.ErrPayrollRunAlreadyExists
		}
		logger.Ctx(ctx).Error("failed to create payroll run", logger.Any("business_id", businessID), logger.Error(err))
		return nil, nil, err
	}

	runID, err := result.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}
	run.ID = runID
	run.OwnerID = businessID
	run.BusinessID = businessID

	createdEntries := make([]domain.PayrollEntryRecord, 0, len(entries))
	for _, entry := range entries {
		entryResult, entryErr := qtx.CreatePayrollEntry(ctx, sqlc.CreatePayrollEntryParams{
			OwnerID:       businessID,
			BusinessID:    businessID,
			PayrollRunID:  runID,
			EmployeeID:    entry.EmployeeID,
			HoursWorked:   fmt.Sprintf("%.2f", entry.HoursWorked),
			OvertimeHours: fmt.Sprintf("%.2f", entry.OvertimeHours),
			BasePay:       centsToDecimalString(entry.BasePayCents),
			OvertimePay:   centsToDecimalString(entry.OvertimePayCents),
			Bonuses:       centsToDecimalString(entry.BonusesCents),
			Deductions:    centsToDecimalString(entry.DeductionsCents),
			NetPay:        centsToDecimalString(entry.NetPayCents),
		})
		if entryErr != nil {
			_ = tx.Rollback()
			logger.Ctx(ctx).Error("failed to create payroll entry", logger.Any("business_id", businessID), logger.Any("employee_id", entry.EmployeeID), logger.Error(entryErr))
			return nil, nil, entryErr
		}

		entryID, idErr := entryResult.LastInsertId()
		if idErr != nil {
			_ = tx.Rollback()
			return nil, nil, idErr
		}
		entry.ID = entryID
		entry.OwnerID = businessID
		entry.BusinessID = businessID
		entry.PayrollRunID = runID
		createdEntries = append(createdEntries, entry)
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	return &run, createdEntries, nil
}
