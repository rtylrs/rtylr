package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/voxire/onion/internal/modules/hr_reports/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		pool: dbclient.Pool(),
	}
}

// appendOptionalFilters adds optional WHERE clauses for location, department, and employee filters.
func appendOptionalFilters(query string, args []interface{}, params domain.ReportParams, employeeAlias string) (string, []interface{}) {
	if params.LocationID != nil {
		query += fmt.Sprintf(" AND %s.location_id = ?", employeeAlias)
		args = append(args, *params.LocationID)
	}
	if params.DepartmentID != nil {
		query += fmt.Sprintf(" AND %s.department_id = ?", employeeAlias)
		args = append(args, *params.DepartmentID)
	}
	if params.EmployeeID != nil {
		query += fmt.Sprintf(" AND %s.id = ?", employeeAlias)
		args = append(args, *params.EmployeeID)
	}
	return query, args
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

// cross-schema aggregation read
func (r *Repository) GenerateHeadcountReport(ctx context.Context, ownerID int64, params domain.ReportParams) ([]map[string]interface{}, error) {
	query := `
		SELECT
			COALESCE(d.name, 'Unassigned') AS department,
			COALESCE(l.name, 'Unassigned') AS location,
			e.employment_type,
			e.status,
			COUNT(*) AS employee_count
		FROM hr_employees e
		LEFT JOIN hr_departments d ON d.id = e.department_id AND d.owner_id = e.owner_id
		LEFT JOIN hr_locations l ON l.id = e.location_id AND l.owner_id = e.owner_id
		WHERE e.owner_id = ?
			AND e.status = 'active'
			AND e.created_at <= ?`

	args := []interface{}{ownerID, params.DateTo}
	query, args = appendOptionalFilters(query, args, params, "e")

	query += " GROUP BY d.name, l.name, e.employment_type, e.status ORDER BY department, location"

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrReportGenerationFailed, err)
	}
	defer rows.Close()

	return scanRows(rows)
}

// cross-schema aggregation read
func (r *Repository) GenerateAttendanceSummaryReport(ctx context.Context, ownerID int64, params domain.ReportParams) ([]map[string]interface{}, error) {
	query := `
		SELECT
			e.id AS employee_id,
			CONCAT(e.first_name, ' ', e.last_name) AS employee_name,
			COALESCE(d.name, 'Unassigned') AS department,
			COUNT(a.id) AS total_records,
			SUM(CASE WHEN a.status = 'present' THEN 1 ELSE 0 END) AS days_present,
			SUM(CASE WHEN a.is_late = TRUE THEN 1 ELSE 0 END) AS days_late,
			COALESCE(SUM(a.late_minutes), 0) AS total_late_minutes,
			COALESCE(SUM(a.duration_minutes), 0) AS total_duration_minutes
		FROM hr_employees e
		LEFT JOIN hr_departments d ON d.id = e.department_id AND d.owner_id = e.owner_id
		LEFT JOIN hr_attendance_records a ON a.employee_id = e.id AND a.owner_id = e.owner_id
			AND a.date BETWEEN ? AND ?
		WHERE e.owner_id = ?
			AND e.status = 'active'`

	args := []interface{}{params.DateFrom, params.DateTo, ownerID}
	query, args = appendOptionalFilters(query, args, params, "e")

	query += " GROUP BY e.id, e.first_name, e.last_name, d.name ORDER BY employee_name"

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrReportGenerationFailed, err)
	}
	defer rows.Close()

	return scanRows(rows)
}

// cross-schema aggregation read
func (r *Repository) GenerateLeaveSummaryReport(ctx context.Context, ownerID int64, params domain.ReportParams) ([]map[string]interface{}, error) {
	query := `
		SELECT
			e.id AS employee_id,
			CONCAT(e.first_name, ' ', e.last_name) AS employee_name,
			COALESCE(d.name, 'Unassigned') AS department,
			lt.name AS leave_type,
			COALESCE(lb.entitled_days, 0) AS entitled_days,
			COALESCE(lb.used_days, 0) AS used_days,
			COALESCE(lb.pending_days, 0) AS pending_days,
			COALESCE(lb.entitled_days - lb.used_days - lb.pending_days, 0) AS remaining_days,
			COUNT(lr.id) AS requests_in_period,
			COALESCE(SUM(CASE WHEN lr.status = 'approved' THEN lr.total_days ELSE 0 END), 0) AS approved_days_in_period
		FROM hr_employees e
		LEFT JOIN hr_departments d ON d.id = e.department_id AND d.owner_id = e.owner_id
		CROSS JOIN hr_leave_types lt ON lt.owner_id = e.owner_id AND lt.is_active = TRUE
		LEFT JOIN hr_leave_balances lb ON lb.employee_id = e.id AND lb.owner_id = e.owner_id
			AND lb.leave_type_id = lt.id AND lb.year = YEAR(?)
		LEFT JOIN hr_leave_requests lr ON lr.employee_id = e.id AND lr.owner_id = e.owner_id
			AND lr.leave_type_id = lt.id
			AND lr.start_date <= ? AND lr.end_date >= ?
		WHERE e.owner_id = ?
			AND e.status = 'active'`

	args := []interface{}{params.DateTo, params.DateTo, params.DateFrom, ownerID}
	query, args = appendOptionalFilters(query, args, params, "e")

	query += " GROUP BY e.id, e.first_name, e.last_name, d.name, lt.id, lt.name, lb.entitled_days, lb.used_days, lb.pending_days ORDER BY employee_name, leave_type"

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrReportGenerationFailed, err)
	}
	defer rows.Close()

	return scanRows(rows)
}

// cross-schema aggregation read
func (r *Repository) GeneratePayrollSummaryReport(ctx context.Context, ownerID int64, params domain.ReportParams) ([]map[string]interface{}, error) {
	query := `
		SELECT
			e.id AS employee_id,
			CONCAT(e.first_name, ' ', e.last_name) AS employee_name,
			COALESCE(d.name, 'Unassigned') AS department,
			COALESCE(SUM(pe.base_pay), 0) AS total_base_pay,
			COALESCE(SUM(pe.overtime_pay), 0) AS total_overtime_pay,
			COALESCE(SUM(pe.bonuses), 0) AS total_bonuses,
			COALESCE(SUM(pe.deductions), 0) AS total_deductions,
			COALESCE(SUM(pe.base_pay + pe.overtime_pay + pe.bonuses), 0) AS total_gross_pay,
			COALESCE(SUM(pe.net_pay), 0) AS total_net_pay,
			COALESCE(SUM(pe.hours_worked), 0) AS total_hours_worked,
			COALESCE(SUM(pe.overtime_hours), 0) AS total_overtime_hours
		FROM hr_employees e
		LEFT JOIN hr_departments d ON d.id = e.department_id AND d.owner_id = e.owner_id
		LEFT JOIN hr_payroll_entries pe ON pe.employee_id = e.id AND pe.owner_id = e.owner_id
		LEFT JOIN hr_payroll_runs pr ON pr.id = pe.payroll_run_id AND pr.owner_id = e.owner_id
			AND pr.period_start >= ? AND pr.period_end <= ?
		WHERE e.owner_id = ?
			AND e.status = 'active'`

	args := []interface{}{params.DateFrom, params.DateTo, ownerID}
	query, args = appendOptionalFilters(query, args, params, "e")

	query += " GROUP BY e.id, e.first_name, e.last_name, d.name ORDER BY employee_name"

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrReportGenerationFailed, err)
	}
	defer rows.Close()

	return scanRows(rows)
}

// cross-schema aggregation read
func (r *Repository) GenerateOvertimeSummaryReport(ctx context.Context, ownerID int64, params domain.ReportParams) ([]map[string]interface{}, error) {
	query := `
		SELECT
			e.id AS employee_id,
			CONCAT(e.first_name, ' ', e.last_name) AS employee_name,
			COALESCE(d.name, 'Unassigned') AS department,
			COUNT(o.id) AS total_requests,
			SUM(CASE WHEN o.status = 'approved' THEN 1 ELSE 0 END) AS approved_requests,
			SUM(CASE WHEN o.status = 'pending' THEN 1 ELSE 0 END) AS pending_requests,
			SUM(CASE WHEN o.status = 'rejected' THEN 1 ELSE 0 END) AS rejected_requests,
			COALESCE(SUM(CASE WHEN o.status = 'approved' THEN o.planned_extra_minutes ELSE 0 END), 0) AS approved_overtime_minutes,
			COALESCE(AVG(CASE WHEN o.status = 'approved' THEN o.multiplier_applied ELSE NULL END), 0) AS avg_multiplier
		FROM hr_employees e
		LEFT JOIN hr_departments d ON d.id = e.department_id AND d.owner_id = e.owner_id
		LEFT JOIN hr_overtime_requests o ON o.employee_id = e.id AND o.owner_id = e.owner_id
			AND o.date BETWEEN ? AND ?
		WHERE e.owner_id = ?
			AND e.status = 'active'`

	args := []interface{}{params.DateFrom, params.DateTo, ownerID}
	query, args = appendOptionalFilters(query, args, params, "e")

	query += " GROUP BY e.id, e.first_name, e.last_name, d.name ORDER BY employee_name"

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrReportGenerationFailed, err)
	}
	defer rows.Close()

	return scanRows(rows)
}

// cross-schema aggregation read
func (r *Repository) GenerateTimesheetSummaryReport(ctx context.Context, ownerID int64, params domain.ReportParams) ([]map[string]interface{}, error) {
	query := `
		SELECT
			e.id AS employee_id,
			CONCAT(e.first_name, ' ', e.last_name) AS employee_name,
			COALESCE(d.name, 'Unassigned') AS department,
			COUNT(DISTINCT ts.id) AS total_timesheets,
			COALESCE(SUM(ts.total_minutes), 0) AS total_logged_minutes,
			COALESCE(SUM(CASE WHEN ts.status = 'approved' THEN ts.total_minutes ELSE 0 END), 0) AS approved_minutes,
			COALESCE(SUM(CASE WHEN ts.status = 'draft' THEN ts.total_minutes ELSE 0 END), 0) AS draft_minutes,
			COALESCE(SUM(CASE WHEN ts.status = 'submitted' THEN ts.total_minutes ELSE 0 END), 0) AS submitted_minutes
		FROM hr_employees e
		LEFT JOIN hr_departments d ON d.id = e.department_id AND d.owner_id = e.owner_id
		LEFT JOIN hr_timesheets ts ON ts.employee_id = e.id AND ts.owner_id = e.owner_id
			AND ts.period_start >= ? AND ts.period_end <= ?
		WHERE e.owner_id = ?
			AND e.status = 'active'`

	args := []interface{}{params.DateFrom, params.DateTo, ownerID}
	query, args = appendOptionalFilters(query, args, params, "e")

	query += " GROUP BY e.id, e.first_name, e.last_name, d.name ORDER BY employee_name"

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrReportGenerationFailed, err)
	}
	defer rows.Close()

	return scanRows(rows)
}
