package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_reports/internal/domain"
)

func TestIsValidReportType_AllValid(t *testing.T) {
	for _, rt := range []string{
		domain.ReportTypeHecount,
		domain.ReportTypeAttendanceSummary,
		domain.ReportTypeLeaveSummary,
		domain.ReportTypePayrollSummary,
		domain.ReportTypeOvertimeSummary,
		domain.ReportTypeTimesheetSummary,
	} {
		if !domain.IsValidReportType(rt) {
			t.Errorf("expected %q to be valid", rt)
		}
	}
}

func TestIsValidReportType_Invalid(t *testing.T) {
	if domain.IsValidReportType("expense_summary") {
		t.Error("expected 'expense_summary' to be invalid")
	}
}

func TestValidateReportParams_Valid(t *testing.T) {
	now := time.Now()
	p := domain.ReportParams{
		OwnerID:  1,
		DateFrom: now.AddDate(0, -1, 0),
		DateTo:   now,
	}
	if err := domain.ValidateReportParams(p); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateReportParams_RejectsInteractiveWindowOver366Days(t *testing.T) {
	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(domain.MaxInteractiveReportWindow + 24*time.Hour)

	err := domain.ValidateReportParams(domain.ReportParams{DateFrom: from, DateTo: to})
	if !errors.Is(err, domain.ErrDateRangeTooLarge) {
		t.Fatalf("expected ErrDateRangeTooLarge, got %v", err)
	}
}

func TestValidateReportParams_ZeroDateFrom(t *testing.T) {
	p := domain.ReportParams{
		OwnerID: 1,
		DateTo:  time.Now(),
	}
	if err := domain.ValidateReportParams(p); !errors.Is(err, domain.ErrMissingDateRange) {
		t.Errorf("expected ErrMissingDateRange, got %v", err)
	}
}

func TestValidateReportParams_ZeroDateTo(t *testing.T) {
	p := domain.ReportParams{
		OwnerID:  1,
		DateFrom: time.Now(),
	}
	if err := domain.ValidateReportParams(p); !errors.Is(err, domain.ErrMissingDateRange) {
		t.Errorf("expected ErrMissingDateRange, got %v", err)
	}
}

func TestValidateReportParams_InvertedRange(t *testing.T) {
	now := time.Now()
	p := domain.ReportParams{
		OwnerID:  1,
		DateFrom: now,
		DateTo:   now.AddDate(0, -1, 0),
	}
	if err := domain.ValidateReportParams(p); !errors.Is(err, domain.ErrDateRangeInverted) {
		t.Errorf("expected ErrDateRangeInverted, got %v", err)
	}
}

func TestValidateReportParams_EqualDates_Rejected(t *testing.T) {
	now := time.Now()
	p := domain.ReportParams{
		OwnerID:  1,
		DateFrom: now,
		DateTo:   now,
	}
	if err := domain.ValidateReportParams(p); !errors.Is(err, domain.ErrDateRangeInverted) {
		t.Errorf("expected ErrDateRangeInverted for equal dates, got %v", err)
	}
}
