package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
)

// fakeWriteGate stands in for the repository-backed permission and ownership
// lookups so the write gate can be exercised without a database.
type fakeWriteGate struct {
	employeeID     int64
	employeeIDErr  error
	granted        map[string]bool
	permissionErr  error
	isOwner        bool
	isOwnerErr     error
	lookupCall     int
	permissionCall int
	ownerCall      int
}

func (f *fakeWriteGate) GetEmployeeIDByUserID(_ context.Context, _, _ int64) (int64, error) {
	f.lookupCall++
	if f.employeeIDErr != nil {
		return 0, f.employeeIDErr
	}
	return f.employeeID, nil
}

func (f *fakeWriteGate) HasPermission(_ context.Context, _, _ int64, permissionKey string) (bool, error) {
	f.permissionCall++
	if f.permissionErr != nil {
		return false, f.permissionErr
	}
	return f.granted[permissionKey], nil
}

func (f *fakeWriteGate) IsBusinessOwner(_ context.Context, _ int64) (bool, error) {
	f.ownerCall++
	if f.isOwnerErr != nil {
		return false, f.isOwnerErr
	}
	return f.isOwner, nil
}

const testCallerUserID = int64(9005)

// The business owner has no hr_employees row and therefore no role. Denying
// them would make the first employee on a fresh tenant uncreatable, because
// the default roles are only seeded during employee creation.
func TestCanManage_BusinessOwnerWithoutEmployeeRowIsAllowed(t *testing.T) {
	gate := &fakeWriteGate{isOwner: true, employeeIDErr: domain.ErrEmployeeNotFound}

	if !canManage(context.Background(), gate, testOwnerID, testCallerUserID, permissionManageEmployees) {
		t.Error("expected the business owner to be allowed to manage employees")
	}
}

// The owner check must short-circuit: an owner is allowed without any employee
// row lookup at all.
func TestCanManage_BusinessOwnerSkipsThePermissionLookup(t *testing.T) {
	gate := &fakeWriteGate{isOwner: true}

	canManage(context.Background(), gate, testOwnerID, testCallerUserID, permissionManageEmployees)

	if gate.lookupCall != 0 || gate.permissionCall != 0 {
		t.Errorf("expected no employee or permission lookup for an owner, got %d and %d",
			gate.lookupCall, gate.permissionCall)
	}
}

func TestCanManage_EmployeeHoldingThePermissionIsAllowed(t *testing.T) {
	gate := &fakeWriteGate{employeeID: 4101, granted: map[string]bool{permissionManageEmployees: true}}

	if !canManage(context.Background(), gate, testOwnerID, testCallerUserID, permissionManageEmployees) {
		t.Error("expected an employee holding manage_employees to be allowed")
	}
}

// The reproduction in issue #62: a plain Employee role rewriting a coworker's
// salary.
func TestCanManage_PlainEmployeeIsDenied(t *testing.T) {
	gate := &fakeWriteGate{employeeID: 4101, granted: map[string]bool{"view_own_profile": true}}

	if canManage(context.Background(), gate, testOwnerID, testCallerUserID, permissionManageEmployees) {
		t.Error("expected a plain employee to be denied")
	}
}

// Permission keys must not be interchangeable: managing employees does not
// imply managing roles, which is what stops a gated caller from granting
// themselves HR Admin.
func TestCanManage_PermissionKeyIsNotInterchangeable(t *testing.T) {
	gate := &fakeWriteGate{employeeID: 4101, granted: map[string]bool{permissionManageEmployees: true}}

	if canManage(context.Background(), gate, testOwnerID, testCallerUserID, permissionManageRoles) {
		t.Error("expected manage_employees not to satisfy a manage_roles gate")
	}
}

func TestCanManage_NonOwnerWithoutEmployeeRowIsDenied(t *testing.T) {
	gate := &fakeWriteGate{employeeIDErr: domain.ErrEmployeeNotFound}

	if canManage(context.Background(), gate, testOwnerID, testCallerUserID, permissionManageEmployees) {
		t.Error("expected a caller with no employee row and no business to be denied")
	}
}

func TestCanManage_FailsClosedWhenPermissionLookupErrors(t *testing.T) {
	gate := &fakeWriteGate{employeeID: 4101, permissionErr: errors.New("db down")}

	if canManage(context.Background(), gate, testOwnerID, testCallerUserID, permissionManageEmployees) {
		t.Error("expected a failed permission lookup to deny the write")
	}
}

// A failed ownership lookup must not grant the write, but must also not lock
// out an HR admin who would pass on their role alone.
func TestCanManage_FallsBackToPermissionWhenOwnershipLookupErrors(t *testing.T) {
	gate := &fakeWriteGate{
		isOwnerErr: errors.New("db down"),
		employeeID: 4101,
		granted:    map[string]bool{permissionManageEmployees: true},
	}

	if !canManage(context.Background(), gate, testOwnerID, testCallerUserID, permissionManageEmployees) {
		t.Error("expected an HR admin to still pass when the ownership lookup fails")
	}
}

func TestCanManage_FailsClosedWhenOwnershipAndPermissionBothError(t *testing.T) {
	gate := &fakeWriteGate{isOwnerErr: errors.New("db down"), employeeIDErr: errors.New("db down")}

	if canManage(context.Background(), gate, testOwnerID, testCallerUserID, permissionManageEmployees) {
		t.Error("expected both lookups failing to deny the write")
	}
}

// An unauthenticated or unresolved caller must never reach the lookups.
func TestCanManage_ZeroCallerUserIDIsDenied(t *testing.T) {
	gate := &fakeWriteGate{isOwner: true}

	if canManage(context.Background(), gate, testOwnerID, 0, permissionManageEmployees) {
		t.Error("expected a zero caller user id to be denied")
	}
	if gate.ownerCall != 0 {
		t.Errorf("expected no ownership lookup for a zero caller id, got %d", gate.ownerCall)
	}
}
