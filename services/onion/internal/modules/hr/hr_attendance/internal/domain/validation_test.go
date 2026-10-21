package domain_test

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_attendance/internal/domain"
)

func TestIsValidClockInMethod_AllValid(t *testing.T) {
	for _, m := range []string{
		domain.ClockInMethodPin,
		domain.ClockInMethodQR,
		domain.ClockInMethodApp,
		domain.ClockInMethodFacial,
		domain.ClockInMethodGPS,
	} {
		if !domain.IsValidClockInMethod(m) {
			t.Errorf("expected %q to be valid", m)
		}
	}
}

func TestIsValidClockInMethod_Invalid(t *testing.T) {
	if domain.IsValidClockInMethod("biometric") {
		t.Error("expected 'biometric' to be invalid")
	}
}

func TestIsValidAttendanceStatus_AllValid(t *testing.T) {
	for _, s := range []string{
		domain.AttendanceStatusClockedIn,
		domain.AttendanceStatusClockedOut,
		domain.AttendanceStatusAbsent,
	} {
		if !domain.IsValidAttendanceStatus(s) {
			t.Errorf("expected %q to be valid", s)
		}
	}
}

func TestIsValidAttendanceStatus_Invalid(t *testing.T) {
	if domain.IsValidAttendanceStatus("on_break") {
		t.Error("expected 'on_break' to be invalid")
	}
}

func TestValidateAttendanceRecord_Valid(t *testing.T) {
	rec := domain.AttendanceRecordRecord{
		ClockInMethod:   domain.ClockInMethodPin,
		Status:          domain.AttendanceStatusClockedIn,
		DurationMinutes: 0,
	}
	if err := domain.ValidateAttendanceRecord(rec); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateAttendanceRecord_InvalidMethod(t *testing.T) {
	rec := domain.AttendanceRecordRecord{
		ClockInMethod: "invalid",
		Status:        domain.AttendanceStatusClockedIn,
	}
	if err := domain.ValidateAttendanceRecord(rec); !errors.Is(err, domain.ErrInvalidClockInMethod) {
		t.Errorf("expected ErrInvalidClockInMethod, got %v", err)
	}
}

func TestValidateAttendanceRecord_InvalidStatus(t *testing.T) {
	rec := domain.AttendanceRecordRecord{
		ClockInMethod: domain.ClockInMethodPin,
		Status:        "on_break",
	}
	if err := domain.ValidateAttendanceRecord(rec); !errors.Is(err, domain.ErrInvalidAttendanceStatus) {
		t.Errorf("expected ErrInvalidAttendanceStatus, got %v", err)
	}
}

func TestValidateAttendanceRecord_NegativeDuration(t *testing.T) {
	rec := domain.AttendanceRecordRecord{
		ClockInMethod:   domain.ClockInMethodApp,
		Status:          domain.AttendanceStatusClockedOut,
		DurationMinutes: -10,
	}
	if err := domain.ValidateAttendanceRecord(rec); !errors.Is(err, domain.ErrNegativeDuration) {
		t.Errorf("expected ErrNegativeDuration, got %v", err)
	}
}
