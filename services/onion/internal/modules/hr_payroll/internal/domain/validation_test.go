package domain_test

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_payroll/internal/domain"
)

func TestIsValidPayrollStatus_AllValid(t *testing.T) {
	for _, s := range []string{
		domain.PayrollStatusDraft,
		domain.PayrollStatusInReview,
		domain.PayrollStatusApproved,
		domain.PayrollStatusExported,
	} {
		if !domain.IsValidPayrollStatus(s) {
			t.Errorf("expected %q to be valid", s)
		}
	}
}

func TestIsValidPayrollStatus_Invalid(t *testing.T) {
	if domain.IsValidPayrollStatus("processing") {
		t.Error("expected 'processing' to be invalid")
	}
}

func TestIsValidPeriodType_AllValid(t *testing.T) {
	for _, pt := range []string{
		domain.PayrollPeriodMonthly,
		domain.PayrollPeriodBiWeekly,
		domain.PayrollPeriodWeekly,
		domain.PayrollPeriodSemiMonthly,
	} {
		if !domain.IsValidPeriodType(pt) {
			t.Errorf("expected %q to be valid", pt)
		}
	}
}

func TestIsValidPeriodType_Invalid(t *testing.T) {
	if domain.IsValidPeriodType("daily") {
		t.Error("expected 'daily' to be invalid")
	}
}

func TestValidatePayrollRun_Valid(t *testing.T) {
	run := domain.PayrollRunRecord{
		PeriodStart:     "2026-01-01",
		PeriodEnd:       "2026-01-31",
		Status:          domain.PayrollStatusDraft,
		TotalGrossCents: 500000,
		TotalNetCents:   400000,
	}
	if err := domain.ValidatePayrollRun(run); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidatePayrollRun_EmptyPeriodStart(t *testing.T) {
	run := domain.PayrollRunRecord{
		PeriodStart: "",
		PeriodEnd:   "2026-01-31",
		Status:      domain.PayrollStatusDraft,
	}
	if err := domain.ValidatePayrollRun(run); !errors.Is(err, domain.ErrEmptyPeriodDates) {
		t.Errorf("expected ErrEmptyPeriodDates, got %v", err)
	}
}

func TestValidatePayrollRun_InvalidStatus(t *testing.T) {
	run := domain.PayrollRunRecord{
		PeriodStart: "2026-01-01",
		PeriodEnd:   "2026-01-31",
		Status:      "cancelled",
	}
	if err := domain.ValidatePayrollRun(run); !errors.Is(err, domain.ErrInvalidPayrollStatus) {
		t.Errorf("expected ErrInvalidPayrollStatus, got %v", err)
	}
}

func TestValidatePayrollRun_NegativeTotalGross(t *testing.T) {
	run := domain.PayrollRunRecord{
		PeriodStart:     "2026-01-01",
		PeriodEnd:       "2026-01-31",
		Status:          domain.PayrollStatusDraft,
		TotalGrossCents: -1,
	}
	if err := domain.ValidatePayrollRun(run); !errors.Is(err, domain.ErrNegativePayAmount) {
		t.Errorf("expected ErrNegativePayAmount, got %v", err)
	}
}

func TestValidatePayrollEntry_Valid(t *testing.T) {
	entry := domain.PayrollEntryRecord{
		BasePayCents:     200000,
		OvertimePayCents: 30000,
		BonusesCents:     10000,
		DeductionsCents:  20000,
		NetPayCents:      220000,
	}
	if err := domain.ValidatePayrollEntry(entry); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidatePayrollEntry_NegativeBasePay(t *testing.T) {
	entry := domain.PayrollEntryRecord{BasePayCents: -1}
	if err := domain.ValidatePayrollEntry(entry); !errors.Is(err, domain.ErrNegativePayAmount) {
		t.Errorf("expected ErrNegativePayAmount, got %v", err)
	}
}

func TestComputeNetPay_CorrectCalculation(t *testing.T) {
	entry := domain.PayrollEntryRecord{
		BasePayCents:     200000,
		OvertimePayCents: 30000,
		BonusesCents:     10000,
		DeductionsCents:  20000,
	}
	got := domain.ComputeNetPay(entry)
	want := int64(220000)
	if got != want {
		t.Errorf("ComputeNetPay = %v, want %v", got, want)
	}
}

func TestComputeNetPay_NoExtras(t *testing.T) {
	entry := domain.PayrollEntryRecord{BasePayCents: 300000}
	if domain.ComputeNetPay(entry) != 300000 {
		t.Error("ComputeNetPay with no extras should equal BasePay")
	}
}

func TestComputeNetPay_ZeroEntry(t *testing.T) {
	entry := domain.PayrollEntryRecord{}
	if domain.ComputeNetPay(entry) != 0 {
		t.Error("ComputeNetPay for zero entry should be 0")
	}
}
