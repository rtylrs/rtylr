package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/hr_reports/internal/domain"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		pool: db,
	}
}

// appendOptionalFilters adds optional WHERE clauses for location, department, and employee filters.
func appendOptionalFilters(builder sq.SelectBuilder, params domain.ReportParams, employeeAlias string) sq.SelectBuilder {
	if params.LocationID != nil {
		builder = builder.Where(sq.Eq{employeeAlias + ".location_id": *params.LocationID})
	}
	if params.DepartmentID != nil {
		builder = builder.Where(sq.Eq{employeeAlias + ".department_id": *params.DepartmentID})
	}
	if params.EmployeeID != nil {
		builder = builder.Where(sq.Eq{employeeAlias + ".id": *params.EmployeeID})
	}
	return builder
}

// scanRows converts sql.Rows into a slice of maps keyed by column name.
func scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if scanErr := rows.Scan(valuePtrs...); scanErr != nil {
			return nil, scanErr
		}

		row := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// cross-schema aggregation read.
func (r *Repository) GenerateHeadcountReport(ctx context.Context, businessID int64, params domain.ReportParams) ([]map[string]interface{}, error) {
	builder := sq.Select(
		"COALESCE(d.name, 'Unassigned') AS department",
		"COALESCE(l.name, 'Unassigned') AS location",
		"e.employment_type",
		"e.status",
		"COUNT(*) AS employee_count",
	).
		From("hr_employees e").
		LeftJoin("hr_departments d ON d.id = e.department_id AND d.business_id = e.business_id").
		LeftJoin("hr_locations l ON l.id = e.location_id AND l.business_id = e.business_id").
		Where(sq.Eq{"e.business_id": businessID}).
		Where(sq.Eq{"e.status": "active"}).
		Where(sq.LtOrEq{"e.created_at": params.DateTo})

	builder = appendOptionalFilters(builder, params, "e")

	builder = builder.
		GroupBy("d.name", "l.name", "e.employment_type", "e.status").
		OrderBy("department", "location").
		PlaceholderFormat(sq.Question)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrReportGenerationFailed, err)
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrReportGenerationFailed, err)
	}
	defer rows.Close()

	return scanRows(rows)
}

// cross-schema aggregation read.
func (r *Repository) GenerateAttendanceSummaryReport(ctx context.Context, businessID int64, params domain.ReportParams) ([]map[string]interface{}, error) {
	builder := sq.Select(
		"e.id AS employee_id",
		"CONCAT(e.first_name, ' ', e.last_name) AS employee_name",
		"COALESCE(d.name, 'Unassigned') AS department",
		"COUNT(a.id) AS total_records",
		"SUM(CASE WHEN a.status = 'present' THEN 1 ELSE 0 END) AS days_present",
		"SUM(CASE WHEN a.is_late = TRUE THEN 1 ELSE 0 END) AS days_late",
		"COALESCE(SUM(a.late_minutes), 0) AS total_late_minutes",
		"COALESCE(SUM(a.duration_minutes), 0) AS total_duration_minutes",
	).
		From("hr_employees e").
		LeftJoin("hr_departments d ON d.id = e.department_id AND d.business_id = e.business_id").
		LeftJoin("hr_attendance_records a ON a.employee_id = e.id AND a.business_id = e.business_id AND a.date BETWEEN ? AND ?", params.DateFrom, params.DateTo).
		Where(sq.Eq{"e.business_id": businessID}).
		Where(sq.Eq{"e.status": "active"})

	builder = appendOptionalFilters(builder, params, "e")

	builder = builder.
		GroupBy("e.id", "e.first_name", "e.last_name", "d.name").
		OrderBy("employee_name").
		PlaceholderFormat(sq.Question)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrReportGenerationFailed, err)
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrReportGenerationFailed, err)
	}
	defer rows.Close()

	return scanRows(rows)
}

// cross-schema aggregation read.
func (r *Repository) GenerateLeaveSummaryReport(ctx context.Context, businessID int64, params domain.ReportParams) ([]map[string]interface{}, error) {
	builder := sq.Select(
		"e.id AS employee_id",
		"CONCAT(e.first_name, ' ', e.last_name) AS employee_name",
		"COALESCE(d.name, 'Unassigned') AS department",
		"lt.name AS leave_type",
		"COALESCE(lb.entitled_days, 0) AS entitled_days",
		"COALESCE(lb.used_days, 0) AS used_days",
		"COALESCE(lb.pending_days, 0) AS pending_days",
		"COALESCE(lb.entitled_days - lb.used_days - lb.pending_days, 0) AS remaining_days",
		"COUNT(lr.id) AS requests_in_period",
		"COALESCE(SUM(CASE WHEN lr.status = 'approved' THEN lr.total_days ELSE 0 END), 0) AS approved_days_in_period",
	).
		From("hr_employees e").
		LeftJoin("hr_departments d ON d.id = e.department_id AND d.business_id = e.business_id").
		CrossJoin("hr_leave_types lt ON lt.business_id = e.business_id AND lt.is_active = TRUE").
		LeftJoin("hr_leave_balances lb ON lb.employee_id = e.id AND lb.business_id = e.business_id AND lb.leave_type_id = lt.id AND lb.year = YEAR(?)", params.DateTo).
		LeftJoin("hr_leave_requests lr ON lr.employee_id = e.id AND lr.business_id = e.business_id AND lr.leave_type_id = lt.id AND lr.start_date <= ? AND lr.end_date >= ?", params.DateTo, params.DateFrom).
		Where(sq.Eq{"e.business_id": businessID}).
		Where(sq.Eq{"e.status": "active"})

	builder = appendOptionalFilters(builder, params, "e")

	builder = builder.
		GroupBy("e.id", "e.first_name", "e.last_name", "d.name", "lt.id", "lt.name", "lb.entitled_days", "lb.used_days", "lb.pending_days").
		OrderBy("employee_name", "leave_type").
		PlaceholderFormat(sq.Question)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrReportGenerationFailed, err)
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrReportGenerationFailed, err)
	}
	defer rows.Close()

	return scanRows(rows)
}

// cross-schema aggregation read.
func (r *Repository) GeneratePayrollSummaryReport(ctx context.Context, businessID int64, params domain.ReportParams) ([]map[string]interface{}, error) {
	builder := sq.Select(
		"e.id AS employee_id",
		"CONCAT(e.first_name, ' ', e.last_name) AS employee_name",
		"COALESCE(d.name, 'Unassigned') AS department",
		"COALESCE(SUM(pe.base_pay), 0) AS total_base_pay",
		"COALESCE(SUM(pe.overtime_pay), 0) AS total_overtime_pay",
		"COALESCE(SUM(pe.bonuses), 0) AS total_bonuses",
		"COALESCE(SUM(pe.deductions), 0) AS total_deductions",
		"COALESCE(SUM(pe.base_pay + pe.overtime_pay + pe.bonuses), 0) AS total_gross_pay",
		"COALESCE(SUM(pe.net_pay), 0) AS total_net_pay",
		"COALESCE(SUM(pe.hours_worked), 0) AS total_hours_worked",
		"COALESCE(SUM(pe.overtime_hours), 0) AS total_overtime_hours",
	).
		From("hr_employees e").
		LeftJoin("hr_departments d ON d.id = e.department_id AND d.business_id = e.business_id").
		LeftJoin("hr_payroll_entries pe ON pe.employee_id = e.id AND pe.business_id = e.business_id").
		LeftJoin("hr_payroll_runs pr ON pr.id = pe.payroll_run_id AND pr.business_id = e.business_id AND pr.period_start >= ? AND pr.period_end <= ?", params.DateFrom, params.DateTo).
		Where(sq.Eq{"e.business_id": businessID}).
		Where(sq.Eq{"e.status": "active"})

	builder = appendOptionalFilters(builder, params, "e")

	builder = builder.
		GroupBy("e.id", "e.first_name", "e.last_name", "d.name").
		OrderBy("employee_name").
		PlaceholderFormat(sq.Question)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrReportGenerationFailed, err)
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrReportGenerationFailed, err)
	}
	defer rows.Close()

	return scanRows(rows)
}

// cross-schema aggregation read.
func (r *Repository) GenerateOvertimeSummaryReport(ctx context.Context, businessID int64, params domain.ReportParams) ([]map[string]interface{}, error) {
	builder := sq.Select(
		"e.id AS employee_id",
		"CONCAT(e.first_name, ' ', e.last_name) AS employee_name",
		"COALESCE(d.name, 'Unassigned') AS department",
		"COUNT(o.id) AS total_requests",
		"SUM(CASE WHEN o.status = 'approved' THEN 1 ELSE 0 END) AS approved_requests",
		"SUM(CASE WHEN o.status = 'pending' THEN 1 ELSE 0 END) AS pending_requests",
		"SUM(CASE WHEN o.status = 'rejected' THEN 1 ELSE 0 END) AS rejected_requests",
		"COALESCE(SUM(CASE WHEN o.status = 'approved' THEN o.planned_extra_minutes ELSE 0 END), 0) AS approved_overtime_minutes",
		"COALESCE(AVG(CASE WHEN o.status = 'approved' THEN o.multiplier_applied ELSE NULL END), 0) AS avg_multiplier",
	).
		From("hr_employees e").
		LeftJoin("hr_departments d ON d.id = e.department_id AND d.business_id = e.business_id").
		LeftJoin("hr_overtime_requests o ON o.employee_id = e.id AND o.business_id = e.business_id AND o.date BETWEEN ? AND ?", params.DateFrom, params.DateTo).
		Where(sq.Eq{"e.business_id": businessID}).
		Where(sq.Eq{"e.status": "active"})

	builder = appendOptionalFilters(builder, params, "e")

	builder = builder.
		GroupBy("e.id", "e.first_name", "e.last_name", "d.name").
		OrderBy("employee_name").
		PlaceholderFormat(sq.Question)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrReportGenerationFailed, err)
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrReportGenerationFailed, err)
	}
	defer rows.Close()

	return scanRows(rows)
}

// cross-schema aggregation read.
func (r *Repository) GenerateTimesheetSummaryReport(ctx context.Context, businessID int64, params domain.ReportParams) ([]map[string]interface{}, error) {
	builder := sq.Select(
		"e.id AS employee_id",
		"CONCAT(e.first_name, ' ', e.last_name) AS employee_name",
		"COALESCE(d.name, 'Unassigned') AS department",
		"COUNT(DISTINCT ts.id) AS total_timesheets",
		"COALESCE(SUM(ts.total_minutes), 0) AS total_logged_minutes",
		"COALESCE(SUM(CASE WHEN ts.status = 'approved' THEN ts.total_minutes ELSE 0 END), 0) AS approved_minutes",
		"COALESCE(SUM(CASE WHEN ts.status = 'draft' THEN ts.total_minutes ELSE 0 END), 0) AS draft_minutes",
		"COALESCE(SUM(CASE WHEN ts.status = 'submitted' THEN ts.total_minutes ELSE 0 END), 0) AS submitted_minutes",
	).
		From("hr_employees e").
		LeftJoin("hr_departments d ON d.id = e.department_id AND d.business_id = e.business_id").
		LeftJoin("hr_timesheets ts ON ts.employee_id = e.id AND ts.business_id = e.business_id AND ts.period_start >= ? AND ts.period_end <= ?", params.DateFrom, params.DateTo).
		Where(sq.Eq{"e.business_id": businessID}).
		Where(sq.Eq{"e.status": "active"})

	builder = appendOptionalFilters(builder, params, "e")

	builder = builder.
		GroupBy("e.id", "e.first_name", "e.last_name", "d.name").
		OrderBy("employee_name").
		PlaceholderFormat(sq.Question)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrReportGenerationFailed, err)
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrReportGenerationFailed, err)
	}
	defer rows.Close()

	return scanRows(rows)
}
