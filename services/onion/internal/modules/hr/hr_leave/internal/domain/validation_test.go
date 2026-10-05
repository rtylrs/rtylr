package domain_test

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/domain"
)

func TestValidateLeaveType_Valid(t *testing.T) {
	lt := domain.LeaveTypeRecord{
		Name:          "Annual Leave",
		AccrualRule:   domain.AccrualRuleMonthly,
		AccrualAmount: 1.5,
		IsActive:      true,
	}
	if err := domain.ValidateLeaveType(lt); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateLeaveType_EmptyName(t *testing.T) {
	lt := domain.LeaveTypeRecord{Name: ""}
	if err := domain.ValidateLeaveType(lt); !errors.Is(err, domain.ErrEmptyLeaveTypeName) {
		t.Errorf("expected ErrEmptyLeaveTypeName, got %v", err)
	}
}

func TestValidateLeaveType_WhitespaceOnlyName(t *testing.T) {
	lt := domain.LeaveTypeRecord{Name: "   "}
	if err := domain.ValidateLeaveType(lt); !errors.Is(err, domain.ErrEmptyLeaveTypeName) {
		t.Errorf("expected ErrEmptyLeaveTypeName for whitespace name, got %v", err)
	}
}

func TestValidateLeaveType_NegativeAccrual(t *testing.T) {
	lt := domain.LeaveTypeRecord{Name: "Sick Leave", AccrualAmount: -1}
	if err := domain.ValidateLeaveType(lt); !errors.Is(err, domain.ErrNegativeAccrual) {
		t.Errorf("expected ErrNegativeAccrual, got %v", err)
	}
}

func TestValidateLeaveType_InvalidAccrualRule(t *testing.T) {
	lt := domain.LeaveTypeRecord{Name: "Sick Leave", AccrualRule: "weekly"}
	if err := domain.ValidateLeaveType(lt); !errors.Is(err, domain.ErrInvalidAccrualRule) {
		t.Errorf("expected ErrInvalidAccrualRule, got %v", err)
	}
}

func TestValidateLeaveType_EmptyAccrualRuleIsOK(t *testing.T) {
	lt := domain.LeaveTypeRecord{Name: "Unpaid Leave", AccrualRule: ""}
	if err := domain.ValidateLeaveType(lt); err != nil {
		t.Errorf("unexpected error for empty accrual rule: %v", err)
	}
}

func TestValidateDateRange_ValidRange(t *testing.T) {
	if err := domain.ValidateDateRange("2026-01-01", "2026-01-05"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateDateRange_SameDay(t *testing.T) {
	if err := domain.ValidateDateRange("2026-03-15", "2026-03-15"); err != nil {
		t.Errorf("same-day range should be valid, got %v", err)
	}
}

func TestValidateDateRange_EndBeforeStart(t *testing.T) {
	if err := domain.ValidateDateRange("2026-05-10", "2026-05-01"); !errors.Is(err, domain.ErrInvalidDateRange) {
		t.Errorf("expected ErrInvalidDateRange, got %v", err)
	}
}

func TestValidateDateRange_InvalidFormat(t *testing.T) {
	if err := domain.ValidateDateRange("not-a-date", "2026-01-01"); !errors.Is(err, domain.ErrInvalidDateRange) {
		t.Errorf("expected ErrInvalidDateRange for bad format, got %v", err)
	}
}

func TestIsValidLeaveStatus_AllValid(t *testing.T) {
	for _, s := range []string{
		domain.LeaveStatusPending,
		domain.LeaveStatusApproved,
		domain.LeaveStatusRejected,
		domain.LeaveStatusCanceled,
	} {
		if !domain.IsValidLeaveStatus(s) {
			t.Errorf("expected %q to be valid", s)
		}
	}
}

func TestIsValidLeaveStatus_Invalid(t *testing.T) {
	if domain.IsValidLeaveStatus("unknown") {
		t.Error("expected 'unknown' to be invalid")
	}
}

func TestIsValidAccrualRule_AllValid(t *testing.T) {
	for _, r := range []string{
		domain.AccrualRuleNone,
		domain.AccrualRuleMonthly,
		domain.AccrualRuleAnnual,
	} {
		if !domain.IsValidAccrualRule(r) {
			t.Errorf("expected %q to be valid", r)
		}
	}
}

func TestIsValidAccrualRule_Invalid(t *testing.T) {
	if domain.IsValidAccrualRule("daily") {
		t.Error("expected 'daily' to be invalid")
	}
}

func TestRemainingBalance_FullBalance(t *testing.T) {
	bal := domain.LeaveBalanceRecord{
		EntitledDays: 20,
		UsedDays:     5,
		PendingDays:  2,
	}
	got := domain.RemainingBalance(bal)
	want := 13.0
	if got != want {
		t.Errorf("RemainingBalance = %v, want %v", got, want)
	}
}

func TestRemainingBalance_ZeroBalance(t *testing.T) {
	bal := domain.LeaveBalanceRecord{
		EntitledDays: 10,
		UsedDays:     10,
		PendingDays:  0,
	}
	if domain.RemainingBalance(bal) != 0 {
		t.Error("expected 0 remaining balance")
	}
}

func TestRemainingBalance_NegativeIsAllowed(t *testing.T) {
	bal := domain.LeaveBalanceRecord{
		EntitledDays: 5,
		UsedDays:     5,
		PendingDays:  3,
	}
	// Pending was added before used ran out; result may be negative
	got := domain.RemainingBalance(bal)
	if got != -3.0 {
		t.Errorf("expected -3.0, got %v", got)
	}
}

func TestValidateLeaveType_RejectsNegativePolicyLimits(t *testing.T) {
	cases := []domain.LeaveTypeRecord{
		{Name: "Annual", MaxCarryoverDays: -1},
		{Name: "Annual", MinNoticeDays: -1},
		{Name: "Annual", RequiresDocumentationAfterDays: -1},
	}
	for _, lt := range cases {
		if err := domain.ValidateLeaveType(lt); !errors.Is(err, domain.ErrNegativeLeavePolicyLimit) {
			t.Errorf("ValidateLeaveType(%+v) error = %v, want ErrNegativeLeavePolicyLimit", lt, err)
		}
	}
}

func TestApplyLeaveTypePatch_PreservesOmittedFields(t *testing.T) {
	current := domain.LeaveTypeRecord{
		Name:                           "Annual Leave",
		IsPaid:                         true,
		RequiresApproval:               true,
		AccrualRule:                    domain.AccrualRuleMonthly,
		AccrualAmount:                  2.5,
		MaxCarryoverDays:               10,
		MinNoticeDays:                  3,
		RequiresDocumentationAfterDays: 7,
		IsActive:                       true,
	}
	name := "Renamed Leave"
	updated := domain.ApplyLeaveTypePatch(current, domain.LeaveTypePatch{Name: &name})
	if updated.Name != name {
		t.Fatalf("name = %q, want %q", updated.Name, name)
	}
	if updated.IsPaid != current.IsPaid || updated.RequiresApproval != current.RequiresApproval ||
		updated.AccrualRule != current.AccrualRule || updated.AccrualAmount != current.AccrualAmount ||
		updated.MaxCarryoverDays != current.MaxCarryoverDays || updated.MinNoticeDays != current.MinNoticeDays ||
		updated.RequiresDocumentationAfterDays != current.RequiresDocumentationAfterDays || updated.IsActive != current.IsActive {
		t.Fatalf("omitted fields changed: got %+v, want all fields except name preserved", updated)
	}
}
