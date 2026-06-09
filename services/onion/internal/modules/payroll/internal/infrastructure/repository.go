package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
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

func (r *Repository) ListPayrollRecords(ctx context.Context, ownerID int64, periodLabel string) ([]domain.PayrollRecord, error) {
	// Dynamic WHERE clause - built with squirrel
	builder := sq.
		Select("id", "owner_id", "staff_id", "period_start", "period_end", "period_label",
			"regular_hours", "overtime_hours", "base_pay", "overtime_pay", "bonuses",
			"deductions", "tax_withheld", "net_pay", "status", "paid_at", "notes", "created_at").
		From("hr_payroll_records").
		Where(sq.Eq{"owner_id": ownerID})

	if periodLabel != "" {
		builder = builder.Where(sq.Eq{"period_label": periodLabel})
	}

	builder = builder.OrderBy("period_start DESC", "staff_id ASC")

	query, args, err := builder.PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build payroll records query", logger.Error(err))
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list payroll records", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var records []domain.PayrollRecord
	for rows.Next() {
		var p domain.PayrollRecord
		if scanErr := p.FromRow(rows); scanErr != nil {
			logger.Ctx(ctx).Error("Failed to scan payroll row", logger.Error(scanErr))
			return nil, scanErr
		}
		records = append(records, p)
	}

	return records, rows.Err()
}

func (r *Repository) CreatePayrollRecord(ctx context.Context, ownerID, staffID int64, periodStart, periodEnd time.Time, periodLabel string, regularHours, overtimeHours, basePay, overtimePay, netPay float64) (int64, error) {
	result, err := r.queries.CreatePayrollRecord(ctx, sqlc.CreatePayrollRecordParams{
		OwnerID:       ownerID,
		StaffID:       staffID,
		PeriodStart:   periodStart,
		PeriodEnd:     periodEnd,
		PeriodLabel:   periodLabel,
		RegularHours:  sql.NullString{String: fmt.Sprintf("%g", regularHours), Valid: true},
		OvertimeHours: sql.NullString{String: fmt.Sprintf("%g", overtimeHours), Valid: true},
		BasePay:       fmt.Sprintf("%g", basePay),
		OvertimePay:   sql.NullString{String: fmt.Sprintf("%g", overtimePay), Valid: true},
		NetPay:        fmt.Sprintf("%g", netPay),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create payroll record", logger.Error(err))
		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) MarkPayrollPaid(ctx context.Context, ownerID, payrollID int64) error {
	result, err := r.queries.MarkPayrollPaid(ctx, sqlc.MarkPayrollPaidParams{
		ID:      payrollID,
		OwnerID: ownerID,
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
func (r *Repository) GetCompletedShiftsForPeriod(ctx context.Context, ownerID int64, periodStart, periodEnd time.Time) ([]domain.ShiftSummary, error) {
	rows, err := r.queries.GetCompletedShiftsForPeriod(ctx, sqlc.GetCompletedShiftsForPeriodParams{
		OwnerID: ownerID,
		Date:    sql.NullTime{Time: periodStart, Valid: true},
		Date_2:  sql.NullTime{Time: periodEnd, Valid: true},
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
func (r *Repository) GetStaffPayRates(ctx context.Context, ownerID, staffID int64) (domain.StaffPayRates, error) {
	row, err := r.queries.GetStaffPayRates(ctx, sqlc.GetStaffPayRatesParams{
		ID:      staffID,
		OwnerID: ownerID,
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
func (r *Repository) GetStaffPayRatesByIDs(ctx context.Context, ownerID int64, staffIDs []int64) (map[int64]domain.StaffPayRates, error) {
	out := make(map[int64]domain.StaffPayRates, len(staffIDs))
	if len(staffIDs) == 0 {
		return out, nil
	}

	query, args, err := sq.
		Select("id", "salary", "hourly_rate", "overtime_rate").
		From("hr_staff").
		Where(sq.Eq{"owner_id": ownerID}).
		Where(sq.Eq{"id": staffIDs}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build staff pay rates query", logger.Error(err))
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to batch-load staff pay rates", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id                               int64
			salary, hourlyRate, overtimeRate sql.NullString
		)
		if scanErr := rows.Scan(&id, &salary, &hourlyRate, &overtimeRate); scanErr != nil {
			return nil, scanErr
		}
		var rates domain.StaffPayRates
		if salary.Valid {
			rates.Salary = parseDecimal(salary.String)
		}
		if hourlyRate.Valid {
			rates.HourlyRate = parseDecimal(hourlyRate.String)
		}
		if overtimeRate.Valid {
			rates.OvertimeRate = parseDecimal(overtimeRate.String)
		}
		out[id] = rates
	}

	return out, rows.Err()
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
