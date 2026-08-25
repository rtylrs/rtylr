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

	"github.com/voxire/onion/internal/modules/payroll/internal/domain"
	"github.com/voxire/onion/internal/modules/payroll/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) CreatePayrollRecords(ctx context.Context, businessID int64, records []domain.PayrollRecordInput) (int, error) {
	if len(records) == 0 {
		return 0, nil
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
	result, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create payroll records", logger.Error(err))
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(affected), nil
}

func (r *Repository) MarkPayrollPaid(ctx context.Context, businessID, payrollID int64) error {
	result, err := r.queries.MarkPayrollPaid(ctx, sqlc.MarkPayrollPaidParams{
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
		return errors.New("payroll record not found or already paid")
	}

	return nil
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
		rates.Salary = parseDecimal(row.Salary.String)
	}
	if row.HourlyRate.Valid {
		rates.HourlyRate = parseDecimal(row.HourlyRate.String)
	}
	if row.OvertimeRate.Valid {
		rates.OvertimeRate = parseDecimal(row.OvertimeRate.String)
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
			rates.Salary = parseDecimal(row.Salary.String)
		}
		if row.HourlyRate.Valid {
			rates.HourlyRate = parseDecimal(row.HourlyRate.String)
		}
		if row.OvertimeRate.Valid {
			rates.OvertimeRate = parseDecimal(row.OvertimeRate.String)
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
		p.RegularHours = parseDecimal(row.RegularHours.String)
	}
	if row.OvertimeHours.Valid {
		p.OvertimeHours = parseDecimal(row.OvertimeHours.String)
	}
	p.BasePay = parseDecimal(row.BasePay)
	if row.OvertimePay.Valid {
		p.OvertimePay = parseDecimal(row.OvertimePay.String)
	}
	if row.Bonuses.Valid {
		p.Bonuses = parseDecimal(row.Bonuses.String)
	}
	if row.Deductions.Valid {
		p.Deductions = parseDecimal(row.Deductions.String)
	}
	if row.TaxWithheld.Valid {
		p.TaxWithheld = parseDecimal(row.TaxWithheld.String)
	}
	p.NetPay = parseDecimal(row.NetPay)
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
func parseDecimal(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse payroll decimal from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}
