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

func TestCanViewSensitiveDetailsRequiresHRPermission(t *testing.T) {
	ctx := context.Background()

	if canViewSensitiveDetails(ctx, fakeStaffPermissionProvider{
		employeeID:  4101,
		permissions: map[string]bool{"view_own_profile": true},
	}, 1001, 700) {
		t.Fatal("ordinary HR user must not receive sensitive detail")
	}

	if !canViewSensitiveDetails(ctx, fakeStaffPermissionProvider{
		employeeID:  4101,
		permissions: map[string]bool{"manage_payroll": true},
	}, 1001, 700) {
		t.Fatal("payroll manager should receive sensitive detail")
	}

	if canViewSensitiveDetails(ctx, fakeStaffPermissionProvider{
		employeeID:    4101,
		permissionErr: errors.New("permission store unavailable"),
	}, 1001, 700) {
		t.Fatal("permission lookup errors must fail closed")
	}
}

func TestNewStaffPayloadOmitsSensitiveDetailWhenUnauthorized(t *testing.T) {
	staff := domain.Staff{
		ID:                    4102,
		FirstName:             "Other",
		LastName:              "Employee",
		Salary:                5000,
		HourlyRate:            30,
		OvertimeRate:          45,
		Address:               "12 Cedar St",
		EmergencyContactName:  "Sam Employee",
		EmergencyContactPhone: "+1-555-0100",
	}

	payload := newStaffPayload(staff, false, 700)
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	var fields map[string]any
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	for _, key := range []string{
		"salary", "hourly_rate", "overtime_rate",
		"address", "emergency_contact_name", "emergency_contact_phone",
	} {
		if _, present := fields[key]; present {
			t.Fatalf("unauthorized payload contains %q: %s", key, encoded)
		}
	}
	if fields["first_name"] != "Other" {
		t.Fatalf("non-sensitive fields were lost: %s", encoded)
	}
}

func TestNewStaffPayloadIncludesSensitiveDetailWhenAuthorized(t *testing.T) {
	payload := newStaffPayload(domain.Staff{
		Salary:                5000,
		HourlyRate:            30,
		OvertimeRate:          45,
		Address:               "12 Cedar St",
		EmergencyContactName:  "Sam Employee",
		EmergencyContactPhone: "+1-555-0100",
	}, true, 0)
	if payload.Salary != 5000 || payload.HourlyRate != 30 || payload.OvertimeRate != 45 {
		t.Fatalf("authorized payload lost compensation: %+v", payload)
	}
	if payload.Address != "12 Cedar St" || payload.EmergencyContactName != "Sam Employee" || payload.EmergencyContactPhone != "+1-555-0100" {
		t.Fatalf("authorized payload lost contact detail: %+v", payload)
	}
}

func TestNewStaffPayloadIncludesSensitiveDetailForOwnRecord(t *testing.T) {
	staff := domain.Staff{
		UserID:                700,
		Salary:                5000,
		HourlyRate:            30,
		OvertimeRate:          45,
		Address:               "12 Cedar St",
		EmergencyContactName:  "Sam Employee",
		EmergencyContactPhone: "+1-555-0100",
	}

	payload := newStaffPayload(staff, false, 700)
	if payload.Salary != 5000 || payload.HourlyRate != 30 || payload.OvertimeRate != 45 {
		t.Fatalf("own record should show compensation even without manage_employees/manage_payroll: %+v", payload)
	}
	if payload.Address != "12 Cedar St" || payload.EmergencyContactName != "Sam Employee" || payload.EmergencyContactPhone != "+1-555-0100" {
		t.Fatalf("own record should show contact detail even without manage_employees/manage_payroll: %+v", payload)
	}
}

func TestNewStaffPayloadOmitsSensitiveDetailWhenCallerUserIDUnresolved(t *testing.T) {
	// A staff record that was never linked to a hub user (UserID left at its
	// zero value) must not be treated as "self" for a caller whose own user
	// id also failed to resolve to 0 - two absent ids are not a match.
	staff := domain.Staff{Salary: 5000}

	payload := newStaffPayload(staff, false, 0)
	if payload.Salary != 0 {
		t.Fatalf("an unresolved caller id must not coincide with an unset staff.UserID: %+v", payload)
	}
}
