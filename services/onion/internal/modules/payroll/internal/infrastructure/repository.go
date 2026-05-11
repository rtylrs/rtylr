package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/payroll/internal/domain"
	"github.com/voxire/onion/internal/modules/payroll/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) ListPayrollRecords(ctx context.Context, ownerID int64, periodLabel string) ([]domain.PayrollRecord, error) {
	// Dynamic WHERE clause - keep on pool
	query := `
		SELECT id, owner_id, staff_id, period_start, period_end, period_label,
		       regular_hours, overtime_hours, base_pay, overtime_pay, bonuses,
		       deductions, tax_withheld, net_pay, status, paid_at, notes, created_at
		FROM hr_payroll_records
		WHERE owner_id = ?`

	args := []interface{}{ownerID}

	if periodLabel != "" {
		query += ` AND period_label = ?`
		args = append(args, periodLabel)
	}

	query += ` ORDER BY period_start DESC, staff_id ASC`

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
	// Aggregation query - keep on pool
	query := `
		SELECT staff_id, SUM(hours_worked) AS total_hours, SUM(overtime_hours) AS total_overtime
		FROM hr_shifts
		WHERE owner_id = ? AND status = 'completed'
		  AND date >= ? AND date <= ?
		GROUP BY staff_id`

	rows, err := r.pool.QueryContext(ctx, query, ownerID,
		periodStart.Format("2006-01-02"),
		periodEnd.Format("2006-01-02"),
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get completed shifts for period", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var summaries []domain.ShiftSummary
	for rows.Next() {
		var s domain.ShiftSummary
		if err := rows.Scan(&s.StaffID, &s.TotalHours, &s.TotalOvertimeHours); err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}

	return summaries, rows.Err()
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
		rates.Salary, _ = strconv.ParseFloat(row.Salary.String, 64)
	}
	if row.HourlyRate.Valid {
		rates.HourlyRate, _ = strconv.ParseFloat(row.HourlyRate.String, 64)
	}
	if row.OvertimeRate.Valid {
		rates.OvertimeRate, _ = strconv.ParseFloat(row.OvertimeRate.String, 64)
	}

	return rates, nil
}
