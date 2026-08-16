package api

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/staff/internal/domain"
)

type fakeStaffPermissionProvider struct {
	employeeID       int64
	lookupErr        error
	permissions      map[string]bool
	permissionErr    error
	directoryEnabled bool
	directoryErr     error
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

func (f fakeStaffPermissionProvider) IsDirectoryEnabled(context.Context, int64) (bool, error) {
	if f.directoryErr != nil {
		return false, f.directoryErr
	}
	return f.directoryEnabled, nil
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

// TestCanViewDirectoryRequiresPermission covers the permission gate on its
// own, so every case here has the business toggle already on. The toggle
// itself is covered by TestCanViewDirectoryDeniesWhenBusinessHasNotEnabledIt.
func TestCanViewDirectoryRequiresPermission(t *testing.T) {
	ctx := context.Background()

	if canViewDirectory(ctx, fakeStaffPermissionProvider{
		directoryEnabled: true,
		employeeID:       4101,
		permissions:      map[string]bool{"manage_payroll": true},
	}, 1001, 700) {
		t.Fatal("holding an unrelated permission must not grant directory access")
	}

	if !canViewDirectory(ctx, fakeStaffPermissionProvider{
		directoryEnabled: true,
		employeeID:       4101,
		permissions:      map[string]bool{"view_directory": true},
	}, 1001, 700) {
		t.Fatal("a caller with view_directory should see the directory")
	}

	if canViewDirectory(ctx, fakeStaffPermissionProvider{
		directoryEnabled: true,
		employeeID:       4101,
		permissionErr:    errors.New("permission store unavailable"),
	}, 1001, 700) {
		t.Fatal("permission lookup errors must fail closed")
	}

	if canViewDirectory(ctx, fakeStaffPermissionProvider{
		directoryEnabled: true,
		lookupErr:        errors.New("employee lookup failed"),
	}, 1001, 700) {
		t.Fatal("an unresolvable caller must fail closed")
	}

	if canViewDirectory(ctx, fakeStaffPermissionProvider{
		directoryEnabled: true,
		employeeID:       4101,
	}, 1001, 0) {
		t.Fatal("a zero caller user id must fail closed")
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

func TestCanViewDirectoryDeniesWhenBusinessHasNotEnabledIt(t *testing.T) {
	ctx := context.Background()
	if canViewDirectory(ctx, fakeStaffPermissionProvider{
		employeeID:       4101,
		directoryEnabled: false,
		permissions:      map[string]bool{"view_directory": true},
	}, 1001, 700) {
		t.Fatal("directory_enabled=false must deny even a caller holding view_directory")
	}
}

func TestCanViewDirectoryDeniesWithoutViewDirectoryPermission(t *testing.T) {
	ctx := context.Background()
	if canViewDirectory(ctx, fakeStaffPermissionProvider{
		employeeID:       4101,
		directoryEnabled: true,
		permissions:      map[string]bool{"view_own_profile": true},
	}, 1001, 700) {
		t.Fatal("a caller without view_directory must be denied even when the business has enabled the directory")
	}
}

func TestCanViewDirectoryAllowsWhenBothGatesPass(t *testing.T) {
	ctx := context.Background()
	if !canViewDirectory(ctx, fakeStaffPermissionProvider{
		employeeID:       4101,
		directoryEnabled: true,
		permissions:      map[string]bool{"view_directory": true},
	}, 1001, 700) {
		t.Fatal("a caller holding view_directory at a business with the directory enabled should be allowed")
	}
}

func TestCanViewDirectoryFailsClosedOnConfigLookupError(t *testing.T) {
	ctx := context.Background()
	if canViewDirectory(ctx, fakeStaffPermissionProvider{
		employeeID:   4101,
		directoryErr: errors.New("hr config store unavailable"),
		permissions:  map[string]bool{"view_directory": true},
	}, 1001, 700) {
		t.Fatal("a directory config lookup error must fail closed")
	}
}

func TestCanViewDirectoryFailsClosedOnPermissionLookupError(t *testing.T) {
	ctx := context.Background()
	if canViewDirectory(ctx, fakeStaffPermissionProvider{
		employeeID:       4101,
		directoryEnabled: true,
		permissionErr:    errors.New("permission store unavailable"),
	}, 1001, 700) {
		t.Fatal("a permission lookup error must fail closed")
	}
}

func TestCanViewDirectoryFailsClosedWithNoResolvableEmployee(t *testing.T) {
	ctx := context.Background()
	if canViewDirectory(ctx, fakeStaffPermissionProvider{
		directoryEnabled: true,
		lookupErr:        errors.New("no employee row for this user"),
		permissions:      map[string]bool{"view_directory": true},
	}, 1001, 700) {
		t.Fatal("a caller with no resolvable employee row must be denied")
	}
}

func TestCanViewDirectoryDeniesNilProvider(t *testing.T) {
	if canViewDirectory(context.Background(), nil, 1001, 700) {
		t.Fatal("a nil permission provider must deny rather than panic")
	}
}

func TestNewDirectoryStaffEntryNeverCarriesSensitiveFields(t *testing.T) {
	entry := newDirectoryStaffEntry(domain.Staff{
		ID:                    4102,
		FirstName:             "Directory",
		LastName:              "Entry",
		Role:                  "Server",
		Salary:                5000,
		Address:               "12 Cedar St",
		EmergencyContactName:  "Sam Employee",
		EmergencyContactPhone: "+1-555-0100",
	})
	encoded, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal entry: %v", err)
	}

	var fields map[string]any
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("unmarshal entry: %v", err)
	}
	for _, key := range []string{
		"salary", "hourly_rate", "overtime_rate",
		"address", "emergency_contact_name", "emergency_contact_phone",
	} {
		if _, present := fields[key]; present {
			t.Fatalf("directory entry contains %q, which DirectoryStaffEntry has no field for: %s", key, encoded)
		}
	}
	if fields["first_name"] != "Directory" || fields["role"] != "Server" {
		t.Fatalf("non-sensitive fields were lost: %s", encoded)
	}
}
