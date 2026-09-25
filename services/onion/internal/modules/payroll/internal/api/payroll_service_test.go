package api

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/voxire/onion/internal/modules/payroll/internal/domain"
)

type fakePayrollPermissionProvider struct {
	employeeID  int64
	permissions map[string]bool
}

func (f fakePayrollPermissionProvider) GetEmployeeIDByUserID(context.Context, int64, int64) (int64, error) {
	return f.employeeID, nil
}

func (f fakePayrollPermissionProvider) HasPermission(_ context.Context, _, _ int64, key string) (bool, error) {
	return f.permissions[key], nil
}

func TestCanViewCompensationFailsClosedForUnprivilegedPayrollCaller(t *testing.T) {
	if canViewCompensation(context.Background(), fakePayrollPermissionProvider{
		employeeID:  4101,
		permissions: map[string]bool{"view_own_profile": true},
	}, 1001, 700) {
		t.Fatal("ordinary HR user must not receive payroll compensation")
	}
}

func TestNewPayrollRecordsOmitsCompensationWhenUnauthorized(t *testing.T) {
	records := newPayrollRecords([]domain.PayrollRecord{{
		ID:            4301,
		StaffID:       4102,
		RegularHours:  40,
		OvertimeHours: 2,
		BasePay:       1000,
		OvertimePay:   150,
		Bonuses:       50,
		Deductions:    100,
		TaxWithheld:   200,
		NetPay:        900,
	}}, false)
	if len(records) != 1 {
		t.Fatalf("got %d payroll records, want 1", len(records))
	}
	if records[0].RegularHours != 40 || records[0].OvertimeHours != 2 {
		t.Fatalf("work hours should remain visible: %+v", records[0])
	}
	for _, value := range []float64{
		records[0].BasePay,
		records[0].OvertimePay,
		records[0].Bonuses,
		records[0].Deductions,
		records[0].TaxWithheld,
		records[0].NetPay,
	} {
		if value != 0 {
			t.Fatalf("unauthorized payroll contains compensation: %+v", records[0])
		}
	}

	encoded, err := json.Marshal(records[0])
	if err != nil {
		t.Fatalf("marshal payroll record: %v", err)
	}
	for _, key := range []string{"base_pay", "overtime_pay", "bonuses", "deductions", "tax_withheld", "net_pay"} {
		if string(encoded) == "" {
			t.Fatal("empty payroll payload")
		}
		if containsJSONField(encoded, key) {
			t.Fatalf("unauthorized payroll contains %q: %s", key, encoded)
		}
	}
}

func TestNewPayrollRecordsIncludesCompensationWhenAuthorized(t *testing.T) {
	records := newPayrollRecords([]domain.PayrollRecord{{BasePay: 1000, NetPay: 900}}, true)
	if records[0].BasePay != 1000 || records[0].NetPay != 900 {
		t.Fatalf("authorized payroll lost compensation: %+v", records[0])
	}
}

func containsJSONField(encoded []byte, key string) bool {
	var fields map[string]any
	if err := json.Unmarshal(encoded, &fields); err != nil {
		return false
	}
	_, present := fields[key]
	return present
}
