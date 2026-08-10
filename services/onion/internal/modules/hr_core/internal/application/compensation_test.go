package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
)

// fakeCompensationViewer stands in for the repository-backed permission lookups
// so the gating logic can be exercised without a database.
type fakeCompensationViewer struct {
	employeeID     int64
	employeeIDErr  error
	granted        map[string]bool
	permissionErr  error
	permissionCall int
	lookupCall     int
}

func (f *fakeCompensationViewer) GetEmployeeIDByUserID(_ context.Context, _, _ int64) (int64, error) {
	f.lookupCall++
	if f.employeeIDErr != nil {
		return 0, f.employeeIDErr
	}
	return f.employeeID, nil
}

func (f *fakeCompensationViewer) HasPermission(_ context.Context, _, _ int64, permissionKey string) (bool, error) {
	f.permissionCall++
	if f.permissionErr != nil {
		return false, f.permissionErr
	}
	return f.granted[permissionKey], nil
}

const (
	testViewerUserID = int64(700)
	testOtherUserID  = int64(800)
	testOwnerID      = int64(1001)
)

func testCoworker() domain.EmployeeRecord {
	return domain.EmployeeRecord{
		ID: 4102, UserID: testOtherUserID, FirstName: "Other",
		Salary: 5000, HourlyRate: 30, OvertimeRate: 45,
	}
}

func testSelf() domain.EmployeeRecord {
	return domain.EmployeeRecord{
		ID: 4101, UserID: testViewerUserID, FirstName: "Self",
		Salary: 3000, HourlyRate: 20, OvertimeRate: 30,
	}
}

func assertCompensationVisible(t *testing.T, e domain.EmployeeRecord, salary, hourly, overtime float64) {
	t.Helper()
	if e.Salary != salary || e.HourlyRate != hourly || e.OvertimeRate != overtime {
		t.Errorf("expected compensation %v/%v/%v, got %v/%v/%v",
			salary, hourly, overtime, e.Salary, e.HourlyRate, e.OvertimeRate)
	}
}

func assertCompensationHidden(t *testing.T, e domain.EmployeeRecord) {
	t.Helper()
	if e.Salary != 0 || e.HourlyRate != 0 || e.OvertimeRate != 0 {
		t.Errorf("expected compensation to be zeroed, got %v/%v/%v",
			e.Salary, e.HourlyRate, e.OvertimeRate)
	}
}

func TestGateCompensation_HRAdminSeesAnotherEmployeeSalary(t *testing.T) {
	viewer := &fakeCompensationViewer{employeeID: 4101, granted: map[string]bool{"manage_employees": true}}

	got := gateCompensation(context.Background(), viewer, testOwnerID, testViewerUserID, []domain.EmployeeRecord{testCoworker()})

	assertCompensationVisible(t, got[0], 5000, 30, 45)
}

func TestGateCompensation_PayrollManagerSeesAnotherEmployeeSalary(t *testing.T) {
	viewer := &fakeCompensationViewer{employeeID: 4101, granted: map[string]bool{"manage_payroll": true}}

	got := gateCompensation(context.Background(), viewer, testOwnerID, testViewerUserID, []domain.EmployeeRecord{testCoworker()})

	assertCompensationVisible(t, got[0], 5000, 30, 45)
}

func TestGateCompensation_PlainEmployeeCannotSeeAnotherEmployeeSalary(t *testing.T) {
	viewer := &fakeCompensationViewer{employeeID: 4101, granted: map[string]bool{"view_own_profile": true}}

	got := gateCompensation(context.Background(), viewer, testOwnerID, testViewerUserID, []domain.EmployeeRecord{testCoworker()})

	assertCompensationHidden(t, got[0])
	if got[0].FirstName != "Other" || got[0].ID != 4102 {
		t.Error("expected non-compensation fields to survive masking")
	}
}

func TestGateCompensation_PlainEmployeeSeesOwnSalary(t *testing.T) {
	viewer := &fakeCompensationViewer{employeeID: 4101, granted: map[string]bool{"view_own_profile": true}}

	got := gateCompensation(context.Background(), viewer, testOwnerID, testViewerUserID, []domain.EmployeeRecord{testSelf()})

	assertCompensationVisible(t, got[0], 3000, 20, 30)
}

func TestGateCompensation_PlainEmployeeInListSeesOnlyOwnRow(t *testing.T) {
	viewer := &fakeCompensationViewer{employeeID: 4101, granted: map[string]bool{"view_own_profile": true}}

	got := gateCompensation(context.Background(), viewer, testOwnerID, testViewerUserID,
		[]domain.EmployeeRecord{testCoworker(), testSelf(), testCoworker()})

	assertCompensationHidden(t, got[0])
	assertCompensationVisible(t, got[1], 3000, 20, 30)
	assertCompensationHidden(t, got[2])
}

// The permission lookup must be resolved once for the caller, not once per row.
func TestGateCompensation_ChecksPermissionsOncePerCallNotPerRow(t *testing.T) {
	viewer := &fakeCompensationViewer{employeeID: 4101, granted: map[string]bool{"view_own_profile": true}}

	rows := make([]domain.EmployeeRecord, 0, 50)
	for i := 0; i < 50; i++ {
		rows = append(rows, testCoworker())
	}
	gateCompensation(context.Background(), viewer, testOwnerID, testViewerUserID, rows)

	if viewer.lookupCall != 1 {
		t.Errorf("expected 1 employee-id lookup, got %d", viewer.lookupCall)
	}
	// One call per compensation key at most: manage_employees, manage_payroll.
	if viewer.permissionCall > len(compensationPermissionKeys) {
		t.Errorf("expected at most %d permission checks, got %d",
			len(compensationPermissionKeys), viewer.permissionCall)
	}
}

// A granted first key must short-circuit the second lookup.
func TestGateCompensation_StopsAtFirstGrantedKey(t *testing.T) {
	viewer := &fakeCompensationViewer{employeeID: 4101, granted: map[string]bool{"manage_employees": true}}

	gateCompensation(context.Background(), viewer, testOwnerID, testViewerUserID, []domain.EmployeeRecord{testCoworker()})

	if viewer.permissionCall != 1 {
		t.Errorf("expected 1 permission check, got %d", viewer.permissionCall)
	}
}

func TestGateCompensation_FailsClosedWhenPermissionLookupErrors(t *testing.T) {
	viewer := &fakeCompensationViewer{employeeID: 4101, permissionErr: errors.New("db down")}

	got := gateCompensation(context.Background(), viewer, testOwnerID, testViewerUserID, []domain.EmployeeRecord{testCoworker()})

	assertCompensationHidden(t, got[0])
}

// A caller with no hr_employees row (e.g. the business owner account) has no
// role and therefore no compensation permission, but must still not be granted
// blanket access.
func TestGateCompensation_FailsClosedWhenCallerHasNoEmployeeRow(t *testing.T) {
	viewer := &fakeCompensationViewer{employeeIDErr: domain.ErrEmployeeNotFound}

	got := gateCompensation(context.Background(), viewer, testOwnerID, testViewerUserID, []domain.EmployeeRecord{testCoworker()})

	assertCompensationHidden(t, got[0])
}

// An unlinked employee row (user_id 0) must never match an unauthenticated or
// zero viewer id via the self-service branch.
func TestGateCompensation_UnlinkedEmployeeIsNotTreatedAsSelf(t *testing.T) {
	viewer := &fakeCompensationViewer{employeeID: 4101, granted: map[string]bool{}}
	unlinked := domain.EmployeeRecord{ID: 4103, UserID: 0, Salary: 900, HourlyRate: 9, OvertimeRate: 12}

	got := gateCompensation(context.Background(), viewer, testOwnerID, 0, []domain.EmployeeRecord{unlinked})

	assertCompensationHidden(t, got[0])
}

func TestGateCompensation_DoesNotMutateInputRecords(t *testing.T) {
	viewer := &fakeCompensationViewer{employeeID: 4101, granted: map[string]bool{}}
	input := []domain.EmployeeRecord{testCoworker()}

	gateCompensation(context.Background(), viewer, testOwnerID, testViewerUserID, input)

	if input[0].Salary != 5000 {
		t.Errorf("expected input record to be left untouched, got salary %v", input[0].Salary)
	}
}
