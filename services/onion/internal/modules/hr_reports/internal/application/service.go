package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/hr_reports/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_reports/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) GenerateReport(ctx context.Context, ownerID int64, reportType string, params domain.ReportParams) ([]map[string]interface{}, error) {
	if !domain.SupportedReportTypes[reportType] {
		return nil, domain.ErrUnsupportedReportType
	}

	switch reportType {
	case domain.ReportTypeHecount:
		return s.repo.GenerateHeadcountReport(ctx, ownerID, params)
	case domain.ReportTypeAttendanceSummary:
		return s.repo.GenerateAttendanceSummaryReport(ctx, ownerID, params)
	case domain.ReportTypeLeaveSummary:
		return s.repo.GenerateLeaveSummaryReport(ctx, ownerID, params)
	case domain.ReportTypePayrollSummary:
		return s.repo.GeneratePayrollSummaryReport(ctx, ownerID, params)
	case domain.ReportTypeOvertimeSummary:
		return s.repo.GenerateOvertimeSummaryReport(ctx, ownerID, params)
	case domain.ReportTypeTimesheetSummary:
		return s.repo.GenerateTimesheetSummaryReport(ctx, ownerID, params)
	default:
		return nil, domain.ErrUnsupportedReportType
	}
}
