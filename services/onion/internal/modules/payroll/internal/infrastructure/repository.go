package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/payroll/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) ListPayrollRecords(ctx context.Context, ownerID int64, periodLabel string) ([]domain.PayrollRecord, error) {
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
	query := `
		INSERT INTO hr_payroll_records
		(owner_id, staff_id, period_start, period_end, period_label, regular_hours, overtime_hours, base_pay, overtime_pay, net_pay)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.pool.ExecContext(ctx, query,
		ownerID, staffID,
		periodStart.Format("2006-01-02"),
		periodEnd.Format("2006-01-02"),
		periodLabel, regularHours, overtimeHours, basePay, overtimePay, netPay,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create payroll record", logger.Error(err))
		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) MarkPayrollPaid(ctx context.Context, ownerID, payrollID int64) error {
	query := `
		UPDATE hr_payroll_records
		SET status = 'paid', paid_at = NOW()
		WHERE id = ? AND owner_id = ? AND status != 'paid'`

	result, err := r.pool.ExecContext(ctx, query, payrollID, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to mark payroll as paid", logger.Error(err))
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return fmt.Errorf("payroll record not found or already paid")
	}

	return nil
}

// GetCompletedShiftsForPeriod retrieves completed shifts for payroll calculation.
func (r *Repository) GetCompletedShiftsForPeriod(ctx context.Context, ownerID int64, periodStart, periodEnd time.Time) ([]domain.ShiftSummary, error) {
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
	var rates domain.StaffPayRates
	var salary, hourlyRate, overtimeRate sql.NullFloat64

	err := r.pool.QueryRowContext(ctx,
		`SELECT salary, hourly_rate, overtime_rate FROM hr_staff WHERE id = ? AND owner_id = ?`,
		staffID, ownerID,
	).Scan(&salary, &hourlyRate, &overtimeRate)
	if err != nil {
		return rates, err
	}

	if salary.Valid {
		rates.Salary = salary.Float64
	}
	if hourlyRate.Valid {
		rates.HourlyRate = hourlyRate.Float64
	}
	if overtimeRate.Valid {
		rates.OvertimeRate = overtimeRate.Float64
	}

	return rates, nil
}
