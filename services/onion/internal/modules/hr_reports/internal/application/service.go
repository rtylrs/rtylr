package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/hr_reports/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_reports/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	hrReportsScope = "hr_reports"
	hrReportsTTL   = 5 * time.Minute
)

type repository interface {
	GenerateHeadcountReport(ctx context.Context, businessID int64, params domain.ReportParams) ([]map[string]interface{}, error)
	GenerateAttendanceSummaryReport(ctx context.Context, businessID int64, params domain.ReportParams) ([]map[string]interface{}, error)
	GenerateLeaveSummaryReport(ctx context.Context, businessID int64, params domain.ReportParams) ([]map[string]interface{}, error)
	GeneratePayrollSummaryReport(ctx context.Context, businessID int64, params domain.ReportParams) ([]map[string]interface{}, error)
	GenerateOvertimeSummaryReport(ctx context.Context, businessID int64, params domain.ReportParams) ([]map[string]interface{}, error)
	GenerateTimesheetSummaryReport(ctx context.Context, businessID int64, params domain.ReportParams) ([]map[string]interface{}, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func paramsKey(p domain.ReportParams) string {
	loc := int64(0)
	if p.LocationID != nil {
		loc = *p.LocationID
	}
	dep := int64(0)
	if p.DepartmentID != nil {
		dep = *p.DepartmentID
	}
	emp := int64(0)
	if p.EmployeeID != nil {
		emp = *p.EmployeeID
	}
	return fmt.Sprintf("f=%d|t=%d|loc=%d|dep=%d|emp=%d",
		p.DateFrom.Unix(), p.DateTo.Unix(), loc, dep, emp)
}

func (s *Service) GenerateReport(ctx context.Context, businessID int64, reportType string, params domain.ReportParams) ([]map[string]interface{}, error) {
	if !domain.SupportedReportTypes[reportType] {
		return nil, domain.ErrUnsupportedReportType
	}

	key := cacheKeys.Scoped(hrReportsScope, businessID, fmt.Sprintf("%s|%s", reportType, paramsKey(params)))
	return cacher.Of[[]map[string]interface{}]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]map[string]interface{}, error) {
			switch reportType {
			case domain.ReportTypeHecount:
				return s.repo.GenerateHeadcountReport(ctx, businessID, params)
			case domain.ReportTypeAttendanceSummary:
				return s.repo.GenerateAttendanceSummaryReport(ctx, businessID, params)
			case domain.ReportTypeLeaveSummary:
				return s.repo.GenerateLeaveSummaryReport(ctx, businessID, params)
			case domain.ReportTypePayrollSummary:
				return s.repo.GeneratePayrollSummaryReport(ctx, businessID, params)
			case domain.ReportTypeOvertimeSummary:
				return s.repo.GenerateOvertimeSummaryReport(ctx, businessID, params)
			case domain.ReportTypeTimesheetSummary:
				return s.repo.GenerateTimesheetSummaryReport(ctx, businessID, params)
			default:
				return nil, domain.ErrUnsupportedReportType
			}
		},
		cacher.WithExpiry(ctx, hrReportsTTL),
	)
}
