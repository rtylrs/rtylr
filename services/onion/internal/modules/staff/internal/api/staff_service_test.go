package api

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/staff/internal/domain"
)

type fakeStaffPermissionProvider struct {
	employeeID    int64
	lookupErr     error
	permissions   map[string]bool
	permissionErr error
}

func (f fakeStaffPermissionProvider) GetEmployeeIDByUserID(context.Context, int64, int64) (int64, error) {
	if f.lookupErr != nil {
		return 0, f.lookupErr
	}
	return f.employeeID, nil
}

func (f fakeStaffPermissionProvider) HasPermission(_ context.Context, _, _ int64, key string) (bool, error) {
	if f.permissionErr != nil {
		return false, f.permissionErr
	}
	return f.permissions[key], nil
}

func TestCanViewCompensationRequiresHRPermission(t *testing.T) {
	ctx := context.Background()

	if canViewCompensation(ctx, fakeStaffPermissionProvider{
		employeeID:  4101,
		permissions: map[string]bool{"view_own_profile": true},
	}, 1001, 700) {
		t.Fatal("ordinary HR user must not receive compensation")
	}

	if !canViewCompensation(ctx, fakeStaffPermissionProvider{
		employeeID:  4101,
		permissions: map[string]bool{"manage_payroll": true},
	}, 1001, 700) {
		t.Fatal("payroll manager should receive compensation")
	}

	if canViewCompensation(ctx, fakeStaffPermissionProvider{
		employeeID:    4101,
		permissionErr: errors.New("permission store unavailable"),
	}, 1001, 700) {
		t.Fatal("permission lookup errors must fail closed")
	}
}

func TestNewStaffPayloadOmitsCompensationWhenUnauthorized(t *testing.T) {
	staff := domain.Staff{
		ID:           4102,
		FirstName:    "Other",
		LastName:     "Employee",
		Salary:       5000,
		HourlyRate:   30,
		OvertimeRate: 45,
	}

	payload := newStaffPayload(staff, false)
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	var fields map[string]any
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	for _, key := range []string{"salary", "hourly_rate", "overtime_rate"} {
		if _, present := fields[key]; present {
			t.Fatalf("unauthorized payload contains %q: %s", key, encoded)
		}
	}
	if fields["first_name"] != "Other" {
		t.Fatalf("non-compensation fields were lost: %s", encoded)
	}
}

func TestNewStaffPayloadIncludesCompensationWhenAuthorized(t *testing.T) {
	payload := newStaffPayload(domain.Staff{Salary: 5000, HourlyRate: 30, OvertimeRate: 45}, true)
	if payload.Salary != 5000 || payload.HourlyRate != 30 || payload.OvertimeRate != 45 {
		t.Fatalf("authorized payload lost compensation: %+v", payload)
	}
}
