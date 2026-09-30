package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/payroll/internal/domain"
	"github.com/voxire/onion/internal/modules/payroll/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) ListPayrollRecords(ctx context.Context, businessID int64, periodLabel string) ([]domain.PayrollRecord, error) {
	var rows []sqlc.HrPayrollRecord
	var err error
	if periodLabel != "" {
		rows, err = r.queries.ListPayrollRecordsByPeriod(ctx, sqlc.ListPayrollRecordsByPeriodParams{
			BusinessID:  businessID,
			PeriodLabel: periodLabel,
		})
	} else {
		rows, err = r.queries.ListPayrollRecords(ctx, businessID)
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list payroll records", logger.Error(err))
		return nil, err
	}

	records := make([]domain.PayrollRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, payrollRecordFromRow(row))
	}
	return records, nil
}

func (r *Repository) CreatePayrollRecord(ctx context.Context, businessID, staffID int64, periodStart, periodEnd time.Time, periodLabel string, regularHours, overtimeHours, basePay, overtimePay, netPay float64) (int64, error) {
	result, err := r.queries.CreatePayrollRecord(ctx, sqlc.CreatePayrollRecordParams{
		OwnerID:       businessID,
		BusinessID:    businessID,
		StaffID:       staffID,
		PeriodStart:   periodStart,
		PeriodEnd:     periodEnd,
		PeriodLabel:   periodLabel,
		RegularHours:  sql.NullString{String: fmt.Sprintf("%g", regularHours), Valid: true},
		OvertimeHours: sql.NullString{String: fmt.Sprintf("%g", overtimeHours), Valid: true},
		BasePay:       fmt.Sprintf("%.2f", basePay),
		OvertimePay:   sql.NullString{String: fmt.Sprintf("%.2f", overtimePay), Valid: true},
		NetPay:        fmt.Sprintf("%.2f", netPay),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create payroll record", logger.Error(err))
		return 0, err
	}

	return result.LastInsertId()
}

// mysqlErrDuplicateEntry is the server error for a unique-key collision, which
// the unique key on (business_id, staff_id, period_label) raises when two runs
// of one period write at the same time.
const mysqlErrDuplicateEntry = 1062

// CreatePayrollRecords replaces one period's pay records inside a single
// transaction, so a second run of the same period lands on the same rows
// instead of adding a second, individually payable copy of them.
//
// A run is refused outright once the period holds a paid record. Recomputing a
// partially paid period would restate figures for money that has already left
// the business and would leave a period whose rows disagree with each other.
func (r *Repository) CreatePayrollRecords(ctx context.Context, businessID int64, records []domain.PayrollRecordInput) (int, error) {
	if len(records) == 0 {
		return 0, nil
	}

	period := records[0]
	for _, record := range records[1:] {
		if !record.PeriodStart.Equal(period.PeriodStart) || !record.PeriodEnd.Equal(period.PeriodEnd) || record.PeriodLabel != period.PeriodLabel {
			return 0, errors.New("payroll records must all cover the same period")
		}
	}

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	qtx := r.queries.WithTx(tx)

	// Lock the period before reading it, so a re-run cannot replace a row that
	// a concurrent settlement is about to mark paid.
	existing, err := qtx.ListPayrollPeriodRowsForUpdate(ctx, sqlc.ListPayrollPeriodRowsForUpdateParams{
		BusinessID:  businessID,
		PeriodStart: period.PeriodStart,
		PeriodEnd:   period.PeriodEnd,
	})
	if err != nil {
		return 0, err
	}
	if domain.PeriodHasPaidRecord(periodRows(existing)) {
		return 0, domain.ErrPayrollPeriodAlreadyPaid
	}

	if err := qtx.DeleteUnpaidPayrollRecords(ctx, sqlc.DeleteUnpaidPayrollRecordsParams{
		BusinessID:  businessID,
		PeriodLabel: period.PeriodLabel,
	}); err != nil {
		return 0, err
	}

	insert := sq.Insert("hr_payroll_records").Columns(
		"owner_id", "business_id", "staff_id", "period_start", "period_end", "period_label",
		"regular_hours", "overtime_hours", "base_pay", "overtime_pay", "net_pay",
	)
	for _, record := range records {
		insert = insert.Values(
			businessID,
			businessID,
			record.StaffID,
			record.PeriodStart,
			record.PeriodEnd,
			record.PeriodLabel,
			fmt.Sprintf("%g", record.RegularHours),
			fmt.Sprintf("%g", record.OvertimeHours),
			fmt.Sprintf("%.2f", record.BasePay),
			fmt.Sprintf("%.2f", record.OvertimePay),
			fmt.Sprintf("%.2f", record.NetPay),
		)
	}
	query, args, err := insert.ToSql()
	if err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlErrDuplicateEntry {
			return 0, domain.ErrPayrollPeriodConflict
		}
		logger.Ctx(ctx).Error("Failed to create payroll records", logger.Error(err))
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return len(records), nil
}

// MarkPayrollPaid settles one payroll record. It refuses when the same staff
// member already has a paid record for the same period: that is the state a
// duplicated run leaves behind, and settling the second row would pay the
// employee twice for one period. A paid record belonging to a different staff
// member is the normal case of paying a period out employee by employee, so it
// is not a conflict.
func (r *Repository) MarkPayrollPaid(ctx context.Context, businessID, payrollID int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	qtx := r.queries.WithTx(tx)

	recordPeriod, err := qtx.GetPayrollRecordPeriod(ctx, sqlc.GetPayrollRecordPeriodParams{
		ID:         payrollID,
		BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrPayrollRecordNotFound
		}
		return err
	}
	periodStart, periodEnd := recordPeriod.PeriodStart, recordPeriod.PeriodEnd

	rows, err := qtx.ListPayrollPeriodRowsForUpdate(ctx, sqlc.ListPayrollPeriodRowsForUpdateParams{
		BusinessID:  businessID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	})
	if err != nil {
		return err
	}

	period := periodRows(rows)

	var staffID int64
	found := false
	for _, row := range period {
		if row.ID != payrollID {
			continue
		}
		found = true
		staffID = row.StaffID
		if row.Status == domain.StatusPaid {
			return domain.ErrPayrollRecordAlreadyPaid
		}
	}
	if !found {
		return domain.ErrPayrollRecordNotFound
	}
	if domain.StaffHasPaidRecordInPeriod(period, staffID, payrollID) {
		return domain.ErrPayrollRecordHasPaidSibling
	}

	result, err := qtx.MarkPayrollPaid(ctx, sqlc.MarkPayrollPaidParams{
		ID:         payrollID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to mark payroll as paid", logger.Error(err))
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrPayrollRecordAlreadyPaid
	}

	return tx.Commit()
}

// GetCompletedShiftsForPeriod retrieves completed shifts for payroll calculation.
func (r *Repository) GetCompletedShiftsForPeriod(ctx context.Context, businessID int64, periodStart, periodEnd time.Time) ([]domain.ShiftSummary, error) {
	rows, err := r.queries.GetCompletedShiftsForPeriod(ctx, sqlc.GetCompletedShiftsForPeriodParams{
		BusinessID: businessID,
		Date:       sql.NullTime{Time: periodStart, Valid: true},
		Date_2:     sql.NullTime{Time: periodEnd, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get completed shifts for period", logger.Error(err))
		return nil, err
	}

	summaries := make([]domain.ShiftSummary, 0, len(rows))
	for _, row := range rows {
		var s domain.ShiftSummary
		if row.StaffID.Valid {
			s.StaffID = row.StaffID.Int64
		}
		s.TotalHours = interfaceToFloat64(row.TotalHours)
		s.TotalOvertimeHours = interfaceToFloat64(row.TotalOvertime)
		summaries = append(summaries, s)
	}

	return summaries, nil
}

// interfaceToFloat64 converts a SUM/AVG aggregate column (which sqlc types as
// interface{}) into a float64, handling the []byte/string forms MySQL returns
// for DECIMAL aggregates.
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

// GetStaffPayRates returns salary/hourly rate for a staff member for payroll calculation.
func (r *Repository) GetStaffPayRates(ctx context.Context, businessID, staffID int64) (domain.StaffPayRates, error) {
	row, err := r.queries.GetStaffPayRates(ctx, sqlc.GetStaffPayRatesParams{
		ID:         staffID,
		BusinessID: businessID,
	})
	if err != nil {
		return domain.StaffPayRates{}, err
	}

	var rates domain.StaffPayRates
	if row.Salary.Valid {
		rates.Salary = sqlutil.ParseDecimal(row.Salary.String)
	}
	if row.HourlyRate.Valid {
		rates.HourlyRate = sqlutil.ParseDecimal(row.HourlyRate.String)
	}
	if row.OvertimeRate.Valid {
		rates.OvertimeRate = sqlutil.ParseDecimal(row.OvertimeRate.String)
	}

	return rates, nil
}

// GetStaffPayRatesByIDs loads pay rates for the given staff IDs in a single
// WHERE id IN (...) query, returning a map keyed by staff ID. This replaces a
// per-staff GetStaffPayRates call (N+1).
func (r *Repository) GetStaffPayRatesByIDs(ctx context.Context, businessID int64, staffIDs []int64) (map[int64]domain.StaffPayRates, error) {
	out := make(map[int64]domain.StaffPayRates, len(staffIDs))
	if len(staffIDs) == 0 {
		return out, nil
	}

	rows, err := r.queries.GetStaffPayRatesByIDs(ctx, sqlc.GetStaffPayRatesByIDsParams{
		BusinessID: businessID,
		StaffIds:   staffIDs,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to batch-load staff pay rates", logger.Error(err))
		return nil, err
	}

	for _, row := range rows {
		var rates domain.StaffPayRates
		if row.Salary.Valid {
			rates.Salary = sqlutil.ParseDecimal(row.Salary.String)
		}
		if row.HourlyRate.Valid {
			rates.HourlyRate = sqlutil.ParseDecimal(row.HourlyRate.String)
		}
		if row.OvertimeRate.Valid {
			rates.OvertimeRate = sqlutil.ParseDecimal(row.OvertimeRate.String)
		}
		out[row.ID] = rates
	}
	return out, nil
}

// GetDefaultOvertimeMultiplier reads the tenant's configured overtime
// multiplier. found is false when hr_overtime has no configuration for the
// business; callers decide what fallback applies.
func (r *Repository) GetDefaultOvertimeMultiplier(ctx context.Context, businessID int64) (float64, bool, error) {
	raw, err := r.queries.GetOvertimeConfigForLegacyPayroll(ctx, businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}

	multiplier, parseErr := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if parseErr != nil {
		logger.Ctx(ctx).Error("Failed to parse overtime multiplier", logger.Error(parseErr))
		return 0, false, parseErr
	}
	return multiplier, true, nil
}

// periodRows narrows a locked period read to what the paid-record decisions in
// this file need, so those decisions stay in the domain where they can be
// tested without a database.
func periodRows(rows []sqlc.ListPayrollPeriodRowsForUpdateRow) []domain.PeriodRow {
	out := make([]domain.PeriodRow, 0, len(rows))
	for _, row := range rows {
		status := ""
		if row.Status.Valid {
			status = string(row.Status.HrPayrollRecordsStatus)
		}
		out = append(out, domain.PeriodRow{
			ID:      row.ID,
			StaffID: row.StaffID,
			Status:  status,
		})
	}
	return out
}

func payrollRecordFromRow(row sqlc.HrPayrollRecord) domain.PayrollRecord {
	var p domain.PayrollRecord
	p.ID = row.ID
	p.OwnerID = row.OwnerID
	p.BusinessID = row.BusinessID
	p.StaffID = row.StaffID
	p.PeriodStart = row.PeriodStart
	p.PeriodEnd = row.PeriodEnd
	p.PeriodLabel = row.PeriodLabel
	if row.RegularHours.Valid {
		p.RegularHours = sqlutil.ParseDecimal(row.RegularHours.String)
	}
	if row.OvertimeHours.Valid {
		p.OvertimeHours = sqlutil.ParseDecimal(row.OvertimeHours.String)
	}
	p.BasePay = sqlutil.ParseDecimal(row.BasePay)
	if row.OvertimePay.Valid {
		p.OvertimePay = sqlutil.ParseDecimal(row.OvertimePay.String)
	}
	if row.Bonuses.Valid {
		p.Bonuses = sqlutil.ParseDecimal(row.Bonuses.String)
	}
	if row.Deductions.Valid {
		p.Deductions = sqlutil.ParseDecimal(row.Deductions.String)
	}
	if row.TaxWithheld.Valid {
		p.TaxWithheld = sqlutil.ParseDecimal(row.TaxWithheld.String)
	}
	p.NetPay = sqlutil.ParseDecimal(row.NetPay)
	p.Status = string(row.Status.HrPayrollRecordsStatus)
	if row.PaidAt.Valid {
		p.PaidAt = row.PaidAt.Time
	}
	if row.Notes.Valid {
		p.Notes = row.Notes.String
	}
	if row.CreatedAt.Valid {
		p.CreatedAt = row.CreatedAt.Time
	}
	return p
}

// parseDecimal converts a DB decimal string to float64, logging corrupt
// values instead of silently coercing pay figures to 0.
