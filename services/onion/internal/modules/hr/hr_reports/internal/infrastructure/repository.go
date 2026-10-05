package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_reports/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_reports/internal/infrastructure/persistence/sqlc"
)

type Repository struct {
	pool    *sql.DB
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		pool:    db,
		queries: sqlc.New(db),
	}
}

func reportFilters(params domain.ReportParams) (sql.NullInt64, sql.NullInt64, sql.NullInt64) {
	return nullInt64Ptr(params.LocationID), nullInt64Ptr(params.DepartmentID), nullInt64Ptr(params.EmployeeID)
}

func nullInt64Ptr(p *int64) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *p, Valid: true}
}

func sqlcRowsToMaps(rows any) ([]map[string]interface{}, error) {
	rv := reflect.ValueOf(rows)
	if rv.Kind() != reflect.Slice {
		return nil, fmt.Errorf("expected slice of report rows")
	}

	results := make([]map[string]interface{}, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		m, err := structRowToMap(rv.Index(i).Interface())
		if err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	return results, nil
}

func structRowToMap(row any) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	rv := reflect.ValueOf(row)
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		name := strings.Split(tag, ",")[0]
		result[name] = reflectValueToInterface(rv.Field(i))
	}
	return result, nil
}

func reflectValueToInterface(v reflect.Value) interface{} {
	switch v.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			return nil
		}
		return reflectValueToInterface(v.Elem())
	case reflect.Pointer:
		if v.IsNil() {
			return nil
		}
		return reflectValueToInterface(v.Elem())
	}

	val := v.Interface()
	switch typed := val.(type) {
	case sql.NullString:
		if typed.Valid {
			return typed.String
		}
		return nil
	case sql.NullInt64:
		if typed.Valid {
			return typed.Int64
		}
		return nil
	case sql.NullFloat64:
		if typed.Valid {
			return typed.Float64
		}
		return nil
	case sql.NullBool:
		if typed.Valid {
			return typed.Bool
		}
		return nil
	case sql.NullTime:
		if typed.Valid {
			return typed.Time
		}
		return nil
	case []byte:
		return string(typed)
	case time.Time:
		return typed
	default:
		return val
	}
}

func wrapReportErr(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %w", domain.ErrReportGenerationFailed, err)
}

// cross-schema aggregation read from hr_core — headcount by department and location.
func (r *Repository) GenerateHeadcountReport(ctx context.Context, businessID int64, params domain.ReportParams) ([]map[string]interface{}, error) {
	loc, dept, emp := reportFilters(params)
	rows, err := r.queries.HeadcountReport(ctx, sqlc.HeadcountReportParams{
		BusinessID:   businessID,
		CreatedAt:    params.DateTo,
		LocationID:   loc,
		DepartmentID: dept,
		EmployeeID:   emp,
	})
	if err != nil {
		return nil, wrapReportErr(err)
	}
	return sqlcRowsToMaps(rows)
}

// cross-schema aggregation read from hr_core + hr_attendance — attendance summary per employee.
func (r *Repository) GenerateAttendanceSummaryReport(ctx context.Context, businessID int64, params domain.ReportParams) ([]map[string]interface{}, error) {
	loc, dept, emp := reportFilters(params)
	rows, err := r.queries.AttendanceSummaryReport(ctx, sqlc.AttendanceSummaryReportParams{
		FromDate:     params.DateFrom,
		ToDate:       params.DateTo,
		BusinessID:   businessID,
		LocationID:   loc,
		DepartmentID: dept,
		EmployeeID:   emp,
	})
	if err != nil {
		return nil, wrapReportErr(err)
	}
	return sqlcRowsToMaps(rows)
}

// cross-schema aggregation read from hr_core + hr_leave — leave balances and requests.
func (r *Repository) GenerateLeaveSummaryReport(ctx context.Context, businessID int64, params domain.ReportParams) ([]map[string]interface{}, error) {
	loc, dept, emp := reportFilters(params)
	rows, err := r.queries.LeaveSummaryReport(ctx, sqlc.LeaveSummaryReportParams{
		YEAR:         params.DateTo,
		EndDate:      params.DateTo,
		StartDate:    params.DateFrom,
		BusinessID:   businessID,
		LocationID:   loc,
		DepartmentID: dept,
		EmployeeID:   emp,
	})
	if err != nil {
		return nil, wrapReportErr(err)
	}
	return sqlcRowsToMaps(rows)
}

// cross-schema aggregation read from hr_core + hr_payroll — payroll totals per employee.
func (r *Repository) GeneratePayrollSummaryReport(ctx context.Context, businessID int64, params domain.ReportParams) ([]map[string]interface{}, error) {
	loc, dept, emp := reportFilters(params)
	rows, err := r.queries.PayrollSummaryReport(ctx, sqlc.PayrollSummaryReportParams{
		PeriodStart:  params.DateFrom,
		PeriodEnd:    params.DateTo,
		BusinessID:   businessID,
		LocationID:   loc,
		DepartmentID: dept,
		EmployeeID:   emp,
	})
	if err != nil {
		return nil, wrapReportErr(err)
	}
	return sqlcRowsToMaps(rows)
}

// cross-schema aggregation read from hr_core + hr_overtime — overtime request summary.
func (r *Repository) GenerateOvertimeSummaryReport(ctx context.Context, businessID int64, params domain.ReportParams) ([]map[string]interface{}, error) {
	loc, dept, emp := reportFilters(params)
	rows, err := r.queries.OvertimeSummaryReport(ctx, sqlc.OvertimeSummaryReportParams{
		FromDate:     params.DateFrom,
		ToDate:       params.DateTo,
		BusinessID:   businessID,
		LocationID:   loc,
		DepartmentID: dept,
		EmployeeID:   emp,
	})
	if err != nil {
		return nil, wrapReportErr(err)
	}
	return sqlcRowsToMaps(rows)
}

// cross-schema aggregation read from hr_core + hr_time — timesheet minutes per employee.
func (r *Repository) GenerateTimesheetSummaryReport(ctx context.Context, businessID int64, params domain.ReportParams) ([]map[string]interface{}, error) {
	loc, dept, emp := reportFilters(params)
	rows, err := r.queries.TimesheetSummaryReport(ctx, sqlc.TimesheetSummaryReportParams{
		PeriodStart:  params.DateFrom,
		PeriodEnd:    params.DateTo,
		BusinessID:   businessID,
		LocationID:   loc,
		DepartmentID: dept,
		EmployeeID:   emp,
	})
	if err != nil {
		return nil, wrapReportErr(err)
	}
	return sqlcRowsToMaps(rows)
}
