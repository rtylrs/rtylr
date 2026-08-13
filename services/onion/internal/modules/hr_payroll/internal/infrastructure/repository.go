package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/hr_payroll/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_payroll/internal/infrastructure/persistence/sqlc"
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
		TotalGross:  fmt.Sprintf("%.2f", run.TotalGross),
		TotalNet:    fmt.Sprintf("%.2f", run.TotalNet),
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
			BasePay:       fmt.Sprintf("%.2f", entry.BasePay),
			OvertimePay:   fmt.Sprintf("%.2f", entry.OvertimePay),
			Bonuses:       fmt.Sprintf("%.2f", entry.Bonuses),
			Deductions:    fmt.Sprintf("%.2f", entry.Deductions),
			NetPay:        fmt.Sprintf("%.2f", entry.NetPay),
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

func (r *Repository) UpdatePayrollRunStatus(ctx context.Context, businessID, runID int64, status string) error {
	return r.queries.UpdatePayrollRunStatus(ctx, sqlc.UpdatePayrollRunStatusParams{
		Status:     status,
		BusinessID: businessID,
		ID:         runID,
	})
}

func (r *Repository) SetPayrollRunApproved(ctx context.Context, businessID, runID, approverID int64) error {
	return r.queries.SetPayrollRunApproved(ctx, sqlc.SetPayrollRunApprovedParams{
		ApprovedBy: sql.NullInt64{Int64: approverID, Valid: true},
		BusinessID: businessID,
		ID:         runID,
	})
}

func (r *Repository) SetPayrollRunExported(ctx context.Context, businessID, runID int64) error {
	return r.queries.SetPayrollRunExported(ctx, sqlc.SetPayrollRunExportedParams{
		BusinessID: businessID,
		ID:         runID,
	})
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

// cross-schema read from hr_overtime — approved overtime minutes with multipliers for payroll computation.
func (r *Repository) ListApprovedOvertimeSummaries(ctx context.Context, ownerID int64, periodStart, periodEnd string) ([]domain.OvertimeSummary, error) {
	ps, err := time.Parse("2006-01-02", periodStart)
	if err != nil {
		return nil, err
	}
	pe, err := time.Parse("2006-01-02", periodEnd)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListApprovedOvertimeSummaries(ctx, sqlc.ListApprovedOvertimeSummariesParams{
		BusinessID: ownerID,
		Date:       ps,
		Date_2:     pe,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list overtime summaries", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, err
	}

	summaries := make([]domain.OvertimeSummary, 0, len(rows))
	for _, row := range rows {
		summaries = append(summaries, domain.OvertimeSummary{
			EmployeeID:   row.EmployeeID,
			TotalMinutes: interfaceToInt32(row.TotalMinutes),
			Multiplier:   interfaceToFloat64(row.AvgMultiplier),
		})
	}
	return summaries, nil
}

// cross-schema read from hr_core — employee contract type, rates, and schedule for payroll computation.
func (r *Repository) ListEmployeePayInfo(ctx context.Context, ownerID int64) ([]domain.EmployeePayInfo, error) {
	rows, err := r.queries.ListEmployeePayInfo(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list employee pay info", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, err
	}

	employees := make([]domain.EmployeePayInfo, 0, len(rows))
	for _, row := range rows {
		hourlyRate := parseFloat(row.HourlyRate.String)
		salary := parseFloat(row.Salary.String)

		payType := "monthly"
		if hourlyRate > 0 && salary == 0 {
			payType = "hourly"
		}

		employees = append(employees, domain.EmployeePayInfo{
			EmployeeID:    row.ID,
			PayType:       payType,
			HourlyRate:    hourlyRate,
			MonthlySalary: salary,
		})
	}
	return employees, nil
}

// cross-schema read from hr_overtime — overtime rate multiplier configuration for payroll computation.
func (r *Repository) GetOvertimeConfig(ctx context.Context, ownerID int64) (float64, bool, error) {
	row, err := r.queries.GetOvertimeConfigForPayroll(ctx, ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return parseFloat(row.DefaultMultiplier), true, nil
}

func (r *Repository) ListDistinctOwnerIDs(ctx context.Context) ([]int64, error) {
	return r.queries.ListDistinctPayrollOwnerIDs(ctx)
}

// ListAllPayrollConfigs returns every payroll config in one query. It exists so
// the period-open notification job can load all owners' configs at once instead
// of listing owner IDs and then fetching each config individually (N+1).
func (r *Repository) ListAllPayrollConfigs(ctx context.Context) ([]domain.PayrollConfigRecord, error) {
	rows, err := r.queries.ListAllPayrollConfigs(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list all payroll configs", logger.Error(err))
		return nil, err
	}

	var configs []domain.PayrollConfigRecord
	for _, row := range rows {
		configs = append(configs, payrollConfigFromRow(row))
	}

	return configs, nil
}

// --- Row Mappers ---

func payrollConfigFromRow(row sqlc.HrPayrollConfig) domain.PayrollConfigRecord {
	return domain.PayrollConfigRecord{
		ID:              row.ID,
		OwnerID:         row.OwnerID,
		PeriodType:      row.PeriodType,
		PayDayOfMonth:   row.PayDayOfMonth,
		OvertimeEnabled: row.OvertimeEnabled,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func payrollRunFromRow(row sqlc.HrPayrollRun) domain.PayrollRunRecord {
	run := domain.PayrollRunRecord{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
		PeriodStart: row.PeriodStart.Format("2006-01-02"),
		PeriodEnd:   row.PeriodEnd.Format("2006-01-02"),
		Status:      row.Status,
		CreatedBy:   row.CreatedBy,
		ApprovedBy:  row.ApprovedBy.Int64,
		TotalGross:  parseFloat(row.TotalGross),
		TotalNet:    parseFloat(row.TotalNet),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
	if row.ExportedAt.Valid {
		run.ExportedAt = row.ExportedAt.Time
	}
	return run
}

func payrollEntryFromRow(row sqlc.HrPayrollEntry) domain.PayrollEntryRecord {
	return domain.PayrollEntryRecord{
		ID:            row.ID,
		OwnerID:       row.OwnerID,
		PayrollRunID:  row.PayrollRunID,
		EmployeeID:    row.EmployeeID,
		HoursWorked:   parseFloat(row.HoursWorked),
		OvertimeHours: parseFloat(row.OvertimeHours),
		BasePay:       parseFloat(row.BasePay),
		OvertimePay:   parseFloat(row.OvertimePay),
		Bonuses:       parseFloat(row.Bonuses),
		Deductions:    parseFloat(row.Deductions),
		NetPay:        parseFloat(row.NetPay),
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

// --- Helpers ---

func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse decimal value from database",
			logger.String("value", s),
			logger.Error(err),
		)
		return 0
	}
	return f
}

func interfaceToInt32(v interface{}) int32 {
	switch val := v.(type) {
	case int64:
		return int32(val)
	case float64:
		return int32(val)
	case []byte:
		n, err := strconv.ParseInt(string(val), 10, 32)
		if err != nil {
			logger.Global().Warn("failed to parse int value from database",
				logger.String("value", string(val)),
				logger.Error(err),
			)
			return 0
		}
		return int32(n)
	case string:
		n, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			logger.Global().Warn("failed to parse int value from database",
				logger.String("value", val),
				logger.Error(err),
			)
			return 0
		}
		return int32(n)
	default:
		return 0
	}
}

func interfaceToFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	case []byte:
		f, err := strconv.ParseFloat(string(val), 64)
		if err != nil {
			logger.Global().Warn("failed to parse float value from database",
				logger.String("value", string(val)),
				logger.Error(err),
			)
			return 0
		}
		return f
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			logger.Global().Warn("failed to parse float value from database",
				logger.String("value", val),
				logger.Error(err),
			)
			return 0
		}
		return f
	default:
		return 0
	}
}
