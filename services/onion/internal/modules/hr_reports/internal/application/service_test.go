package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_reports/internal/domain"
)

type mockRepo struct {
	calledType string
	result     []map[string]interface{}
	err        error
}

func (m *mockRepo) GenerateHeadcountReport(_ context.Context, _ int64, _ domain.ReportParams) ([]map[string]interface{}, error) {
	m.calledType = "headcount"
	return m.result, m.err
}
func (m *mockRepo) GenerateAttendanceSummaryReport(_ context.Context, _ int64, _ domain.ReportParams) ([]map[string]interface{}, error) {
	m.calledType = "attendance"
	return m.result, m.err
}
func (m *mockRepo) GenerateLeaveSummaryReport(_ context.Context, _ int64, _ domain.ReportParams) ([]map[string]interface{}, error) {
	m.calledType = "leave"
	return m.result, m.err
}
func (m *mockRepo) GeneratePayrollSummaryReport(_ context.Context, _ int64, _ domain.ReportParams) ([]map[string]interface{}, error) {
	m.calledType = "payroll"
	return m.result, m.err
}
func (m *mockRepo) GenerateOvertimeSummaryReport(_ context.Context, _ int64, _ domain.ReportParams) ([]map[string]interface{}, error) {
	m.calledType = "overtime"
	return m.result, m.err
}
func (m *mockRepo) GenerateTimesheetSummaryReport(_ context.Context, _ int64, _ domain.ReportParams) ([]map[string]interface{}, error) {
	m.calledType = "timesheet"
	return m.result, m.err
}

func TestGenerateReport_RoutesByType(t *testing.T) {
	cases := []struct {
		typ  string
		want string
	}{
		{domain.ReportTypeHecount, "headcount"},
		{domain.ReportTypeAttendanceSummary, "attendance"},
		{domain.ReportTypeLeaveSummary, "leave"},
		{domain.ReportTypePayrollSummary, "payroll"},
		{domain.ReportTypeOvertimeSummary, "overtime"},
		{domain.ReportTypeTimesheetSummary, "timesheet"},
	}
	for _, tc := range cases {
		m := &mockRepo{}
		s := &Service{repo: m}
		if _, err := s.GenerateReport(context.Background(), 1, tc.typ, domain.ReportParams{}); err != nil {
			t.Errorf("%s: %v", tc.typ, err)
		}
		if m.calledType != tc.want {
			t.Errorf("%s routed to %q, want %q", tc.typ, m.calledType, tc.want)
		}
	}
}

func TestGenerateReport_UnsupportedType(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	_, err := s.GenerateReport(context.Background(), 1, "made_up", domain.ReportParams{})
	if !errors.Is(err, domain.ErrUnsupportedReportType) {
		t.Errorf("expected ErrUnsupportedReportType, got %v", err)
	}
}
