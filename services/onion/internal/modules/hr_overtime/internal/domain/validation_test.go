package domain_test

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_overtime/internal/domain"
)

func TestIsValidOvertimeStatus_AllValid(t *testing.T) {
	for _, s := range []string{
		domain.OvertimeStatusPending,
		domain.OvertimeStatusApproved,
		domain.OvertimeStatusRejected,
	} {
		if !domain.IsValidOvertimeStatus(s) {
			t.Errorf("expected %q to be valid", s)
		}
	}
}

func TestIsValidOvertimeStatus_Invalid(t *testing.T) {
	if domain.IsValidOvertimeStatus("cancelled") {
		t.Error("expected 'cancelled' to be invalid")
	}
}

func TestValidateOvertimeRequest_Valid(t *testing.T) {
	req := domain.OvertimeRequestRecord{
		Date:                "2026-05-20",
		PlannedExtraMinutes: 60,
		Status:              domain.OvertimeStatusPending,
	}
	if err := domain.ValidateOvertimeRequest(req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateOvertimeRequest_EmptyDate(t *testing.T) {
	req := domain.OvertimeRequestRecord{
		Date:                "",
		PlannedExtraMinutes: 60,
		Status:              domain.OvertimeStatusPending,
	}
	if err := domain.ValidateOvertimeRequest(req); !errors.Is(err, domain.ErrEmptyOvertimeDate) {
		t.Errorf("expected ErrEmptyOvertimeDate, got %v", err)
	}
}

func TestValidateOvertimeRequest_ZeroMinutes(t *testing.T) {
	req := domain.OvertimeRequestRecord{
		Date:                "2026-05-20",
		PlannedExtraMinutes: 0,
		Status:              domain.OvertimeStatusPending,
	}
	if err := domain.ValidateOvertimeRequest(req); !errors.Is(err, domain.ErrNegativeExtraMinutes) {
		t.Errorf("expected ErrNegativeExtraMinutes for 0 minutes, got %v", err)
	}
}

func TestValidateOvertimeRequest_NegativeMinutes(t *testing.T) {
	req := domain.OvertimeRequestRecord{
		Date:                "2026-05-20",
		PlannedExtraMinutes: -30,
		Status:              domain.OvertimeStatusPending,
	}
	if err := domain.ValidateOvertimeRequest(req); !errors.Is(err, domain.ErrNegativeExtraMinutes) {
		t.Errorf("expected ErrNegativeExtraMinutes, got %v", err)
	}
}

func TestValidateOvertimeRequest_InvalidStatus(t *testing.T) {
	req := domain.OvertimeRequestRecord{
		Date:                "2026-05-20",
		PlannedExtraMinutes: 60,
		Status:              "unknown",
	}
	if err := domain.ValidateOvertimeRequest(req); !errors.Is(err, domain.ErrInvalidOvertimeStatus) {
		t.Errorf("expected ErrInvalidOvertimeStatus, got %v", err)
	}
}

func TestValidateOvertimeConfig_Valid(t *testing.T) {
	cfg := domain.OvertimeConfigRecord{DefaultMultiplier: 1.5}
	if err := domain.ValidateOvertimeConfig(cfg); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateOvertimeConfig_ZeroMultiplier(t *testing.T) {
	cfg := domain.OvertimeConfigRecord{DefaultMultiplier: 0}
	if err := domain.ValidateOvertimeConfig(cfg); !errors.Is(err, domain.ErrNegativeMultiplier) {
		t.Errorf("expected ErrNegativeMultiplier for 0 multiplier, got %v", err)
	}
}

func TestValidateOvertimeConfig_NegativeMultiplier(t *testing.T) {
	cfg := domain.OvertimeConfigRecord{DefaultMultiplier: -1}
	if err := domain.ValidateOvertimeConfig(cfg); !errors.Is(err, domain.ErrNegativeMultiplier) {
		t.Errorf("expected ErrNegativeMultiplier, got %v", err)
	}
}
