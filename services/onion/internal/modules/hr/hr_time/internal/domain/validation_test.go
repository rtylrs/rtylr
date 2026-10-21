package domain_test

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
)

func TestIsValidEntryStatus_AllValid(t *testing.T) {
	for _, s := range []string{
		domain.EntryStatusDraft,
		domain.EntryStatusApproved,
		domain.EntryStatusRejected,
	} {
		if !domain.IsValidEntryStatus(s) {
			t.Errorf("expected %q to be valid", s)
		}
	}
}

func TestIsValidEntryStatus_Invalid(t *testing.T) {
	if domain.IsValidEntryStatus("pending") {
		t.Error("expected 'pending' to be invalid")
	}
}

func TestIsValidTimesheetStatus_AllValid(t *testing.T) {
	for _, s := range []string{
		domain.TimesheetStatusDraft,
		domain.TimesheetStatusSubmitted,
		domain.TimesheetStatusApproved,
		domain.TimesheetStatusRejected,
	} {
		if !domain.IsValidTimesheetStatus(s) {
			t.Errorf("expected %q to be valid", s)
		}
	}
}

func TestIsValidTimesheetStatus_Invalid(t *testing.T) {
	if domain.IsValidTimesheetStatus("cancelled") {
		t.Error("expected 'cancelled' to be invalid")
	}
}

func TestIsValidEntrySource_AllValid(t *testing.T) {
	for _, s := range []string{
		domain.EntrySourceManual,
		domain.EntrySourceFlow,
		domain.EntrySourceImport,
	} {
		if !domain.IsValidEntrySource(s) {
			t.Errorf("expected %q to be valid", s)
		}
	}
}

func TestIsValidEntrySource_Invalid(t *testing.T) {
	if domain.IsValidEntrySource("csv") {
		t.Error("expected 'csv' to be invalid")
	}
}

func TestValidateTimeEntry_Valid(t *testing.T) {
	e := domain.TimeEntryRecord{
		Date:            "2026-05-01",
		DurationMinutes: 480,
		Status:          domain.EntryStatusDraft,
	}
	if err := domain.ValidateTimeEntry(e); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateTimeEntry_EmptyDate(t *testing.T) {
	e := domain.TimeEntryRecord{Date: "", DurationMinutes: 60}
	if err := domain.ValidateTimeEntry(e); !errors.Is(err, domain.ErrEmptyEntryDate) {
		t.Errorf("expected ErrEmptyEntryDate, got %v", err)
	}
}

func TestValidateTimeEntry_NegativeDuration(t *testing.T) {
	e := domain.TimeEntryRecord{Date: "2026-05-01", DurationMinutes: -1}
	if err := domain.ValidateTimeEntry(e); !errors.Is(err, domain.ErrNegativeEntryDuration) {
		t.Errorf("expected ErrNegativeEntryDuration, got %v", err)
	}
}

func TestValidateTimeEntry_InvalidStatus(t *testing.T) {
	e := domain.TimeEntryRecord{Date: "2026-05-01", DurationMinutes: 60, Status: "unknown"}
	if err := domain.ValidateTimeEntry(e); !errors.Is(err, domain.ErrInvalidEntryStatus) {
		t.Errorf("expected ErrInvalidEntryStatus, got %v", err)
	}
}

func TestValidateTimeEntry_EmptyStatusIsAllowed(t *testing.T) {
	e := domain.TimeEntryRecord{Date: "2026-05-01", DurationMinutes: 30, Status: ""}
	if err := domain.ValidateTimeEntry(e); err != nil {
		t.Errorf("empty status should be allowed, got %v", err)
	}
}

func TestValidateTimesheet_Valid(t *testing.T) {
	ts := domain.TimesheetRecord{
		PeriodStart: "2026-05-01",
		PeriodEnd:   "2026-05-31",
		Status:      domain.TimesheetStatusDraft,
	}
	if err := domain.ValidateTimesheet(ts); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateTimesheet_EmptyPeriodStart(t *testing.T) {
	ts := domain.TimesheetRecord{PeriodStart: "", PeriodEnd: "2026-05-31"}
	if err := domain.ValidateTimesheet(ts); !errors.Is(err, domain.ErrEmptyPeriodDates) {
		t.Errorf("expected ErrEmptyPeriodDates, got %v", err)
	}
}

func TestValidateTimesheet_InvalidStatus(t *testing.T) {
	ts := domain.TimesheetRecord{
		PeriodStart: "2026-05-01",
		PeriodEnd:   "2026-05-31",
		Status:      "invalid_status",
	}
	if err := domain.ValidateTimesheet(ts); !errors.Is(err, domain.ErrInvalidTimesheetStatus) {
		t.Errorf("expected ErrInvalidTimesheetStatus, got %v", err)
	}
}
