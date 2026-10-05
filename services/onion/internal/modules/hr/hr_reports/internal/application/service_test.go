package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_reports/internal/domain"
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

func TestGenerateReport_EmptyTypeIsUnsupported(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	_, err := s.GenerateReport(context.Background(), 1, "", domain.ReportParams{})
	if !errors.Is(err, domain.ErrUnsupportedReportType) {
		t.Errorf("expected ErrUnsupportedReportType for empty type, got %v", err)
	}
}

func TestGenerateReport_PropagatesRepoError(t *testing.T) {
	want := errors.New("db failure")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.GenerateReport(context.Background(), 1, domain.ReportTypeHecount, domain.ReportParams{})
	if !errors.Is(err, want) {
		t.Errorf("expected repo error to propagate, got %v", err)
	}
}

func TestGenerateReport_ReturnsRows(t *testing.T) {
	rows := []map[string]interface{}{{"department": "Engineering", "count": 5}}
	s := &Service{repo: &mockRepo{result: rows}}
	got, err := s.GenerateReport(context.Background(), 1, domain.ReportTypeHecount, domain.ReportParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 row, got %d", len(got))
	}
}

func TestGenerateReport_ReturnsEmptySlice(t *testing.T) {
	s := &Service{repo: &mockRepo{result: []map[string]interface{}{}}}
	got, err := s.GenerateReport(context.Background(), 1, domain.ReportTypeLeaveSummary, domain.ReportParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d rows", len(got))
	}
}

func TestGenerateReport_HeadcountRouting(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	_, _ = s.GenerateReport(context.Background(), 1, domain.ReportTypeHecount, domain.ReportParams{})
	if m.calledType != "headcount" {
		t.Errorf("expected headcount called, got %q", m.calledType)
	}
}

func TestGenerateReport_PayrollRoutingPropagatesError(t *testing.T) {
	want := errors.New("payroll query error")
	m := &mockRepo{err: want}
	s := &Service{repo: m}
	_, err := s.GenerateReport(context.Background(), 1, domain.ReportTypePayrollSummary, domain.ReportParams{})
	if !errors.Is(err, want) {
		t.Errorf("expected payroll repo error, got %v", err)
	}
}

func TestSupportedReportTypes_ContainsAllExpected(t *testing.T) {
	expected := []string{
		domain.ReportTypeHecount,
		domain.ReportTypeAttendanceSummary,
		domain.ReportTypeLeaveSummary,
		domain.ReportTypePayrollSummary,
		domain.ReportTypeOvertimeSummary,
		domain.ReportTypeTimesheetSummary,
	}
	for _, rt := range expected {
		if !domain.SupportedReportTypes[rt] {
			t.Errorf("expected %q to be in SupportedReportTypes", rt)
		}
	}
}

func TestGenerateReport_TimesheetRoutingPropagatesError(t *testing.T) {
	want := errors.New("timesheet error")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.GenerateReport(context.Background(), 1, domain.ReportTypeTimesheetSummary, domain.ReportParams{})
	if !errors.Is(err, want) {
		t.Errorf("expected timesheet repo error to propagate, got %v", err)
	}
}

func TestGenerateReport_OvertimeRoutingPropagatesError(t *testing.T) {
	want := errors.New("overtime error")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.GenerateReport(context.Background(), 1, domain.ReportTypeOvertimeSummary, domain.ReportParams{})
	if !errors.Is(err, want) {
		t.Errorf("expected overtime repo error to propagate, got %v", err)
	}
}

func TestGenerateReport_CaseSensitive(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	_, err := s.GenerateReport(context.Background(), 1, "Headcount", domain.ReportParams{})
	if !errors.Is(err, domain.ErrUnsupportedReportType) {
		t.Errorf("report type matching should be case-sensitive, got %v", err)
	}
}

func TestGenerateReport_RejectsRowsBeyondInteractiveLimit(t *testing.T) {
	rows := make([]map[string]interface{}, maxReportRows+1)
	s := &Service{repo: &mockRepo{result: rows}}

	_, err := s.GenerateReport(context.Background(), 1, domain.ReportTypeHecount, domain.ReportParams{})
	if !errors.Is(err, domain.ErrReportTooLarge) {
		t.Fatalf("expected ErrReportTooLarge, got %v", err)
	}
}
