package workspaces_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces"
	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
	hrleave "github.com/voxire/onion/internal/modules/hr/hr_leave"
	"github.com/voxire/onion/internal/testutil"
)

// Distinct ids so these fixture rows cannot collide with another test's data.
const (
	presenceOwnerID      int64 = 930100
	presenceBusinessID   int64 = 930200
	presenceCallerUserID int64 = 930001 // holds view_directory
	presenceOnsiteUserID int64 = 930002 // has a WorkDay for the test date
	presenceLeaveUserID  int64 = 930003 // has an approved leave request covering the test date
	presenceIdleUserID   int64 = 930004 // neither - should read not_started
	presenceDate               = "2027-05-10"
)

func insertPresenceEmployee(t *testing.T, db *sql.DB, userID int64, firstName string) int64 {
	t.Helper()

	return insertPresenceEmployeeFor(t, db, presenceOwnerID, presenceBusinessID, userID, firstName)
}

// insertPresenceEmployeeFor takes the tenant explicitly so a test can seed a
// second one. Without that, every row belongs to the same business and a
// tenant-scoped read cannot be distinguished from an unscoped one.
func insertPresenceEmployeeFor(
	t *testing.T, db *sql.DB, ownerID, businessID, userID int64, firstName string,
) int64 {
	t.Helper()
	ctx := context.Background()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO hr_employees (owner_id, business_id, user_id, first_name, last_name, email, status, employment_type)
		 VALUES (?, ?, ?, ?, 'Presence', ?, 'active', 'full_time')`,
		ownerID, businessID, userID, firstName, fmt.Sprintf("%s.presence@example.test", firstName),
	); err != nil {
		t.Fatalf("insert hr_employees(%s): %v", firstName, err)
	}
	var employeeID int64
	if err := db.QueryRowContext(ctx,
		`SELECT id FROM hr_employees WHERE owner_id = ? AND business_id = ? AND user_id = ?`,
		ownerID, businessID, userID,
	).Scan(&employeeID); err != nil {
		t.Fatalf("read back hr_employees id(%s): %v", firstName, err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_employees WHERE id = ?", employeeID)
	})
	return employeeID
}

type presenceTestAdapter struct {
	core  *hrcore.Module
	leave *hrleave.Module
}

func (a presenceTestAdapter) EmployeeIDByUserID(ctx context.Context, businessID, userID int64) (int64, error) {
	return a.core.GetEmployeeIDByUserID(ctx, businessID, userID)
}

func (a presenceTestAdapter) HasPermission(ctx context.Context, businessID, employeeID int64, permission string) (bool, error) {
	return a.core.HasPermission(ctx, businessID, employeeID, permission)
}

func (a presenceTestAdapter) ListActiveEmployees(ctx context.Context, businessID int64) ([]workspaces.PresenceEmployee, error) {
	employees, err := a.core.ListActiveEmployees(ctx, businessID)
	if err != nil {
		return nil, err
	}
	result := make([]workspaces.PresenceEmployee, 0, len(employees))
	for _, employee := range employees {
		result = append(result, workspaces.PresenceEmployee{
			ID: employee.ID, UserID: employee.UserID, FirstName: employee.FirstName, LastName: employee.LastName,
		})
	}
	return result, nil
}

func (a presenceTestAdapter) ApprovedLeaveEmployeeIDs(ctx context.Context, userID int64, date string) (map[int64]bool, error) {
	return a.leave.ApprovedLeaveEmployeeIDs(ctx, userID, date)
}

func newPresenceModule(ctx context.Context, db *sql.DB) *workspaces.Module {
	core := hrcore.NewModule(ctx, db)
	leave := hrleave.NewModule(ctx, db, core)
	module := workspaces.NewModule(ctx, db)
	module.SetPresenceSource(presenceTestAdapter{core: core, leave: leave})
	return module
}

func TestListPresenceForBusiness_ResolvesStatusPerEmployee(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	wsModule := newPresenceModule(ctx, db)

	callerEmployeeID := insertPresenceEmployee(t, db, presenceCallerUserID, "Caller")
	onsiteEmployeeID := insertPresenceEmployee(t, db, presenceOnsiteUserID, "Onsite")
	leaveEmployeeID := insertPresenceEmployee(t, db, presenceLeaveUserID, "OnLeave")
	_ = insertPresenceEmployee(t, db, presenceIdleUserID, "Idle")

	// Grant the caller view_directory via a real role + assignment, the same
	// tables SeedDefaultRoles/AssignRole write to - proving the permission
	// path works against the real schema, not a fake.
	if _, err := db.ExecContext(ctx,
		`INSERT INTO hr_roles (owner_id, business_id, name, description, permissions, is_system)
		 VALUES (?, ?, 'Presence Test Role', '', '{"view_directory": true}', false)`,
		presenceOwnerID, presenceBusinessID,
	); err != nil {
		t.Fatalf("insert hr_roles: %v", err)
	}
	var roleID int64
	if err := db.QueryRowContext(ctx,
		`SELECT id FROM hr_roles WHERE business_id = ? AND name = 'Presence Test Role'`, presenceBusinessID,
	).Scan(&roleID); err != nil {
		t.Fatalf("read back hr_roles id: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), "DELETE FROM hr_roles WHERE id = ?", roleID) })
	if _, err := db.ExecContext(ctx,
		`INSERT INTO hr_employee_roles (owner_id, business_id, employee_id, role_id) VALUES (?, ?, ?, ?)`,
		presenceOwnerID, presenceBusinessID, callerEmployeeID, roleID,
	); err != nil {
		t.Fatalf("insert hr_employee_roles: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_employee_roles WHERE role_id = ?", roleID)
	})

	if _, err := db.ExecContext(ctx,
		`INSERT INTO flow_work_days (owner_id, business_id, employee_id, date, timezone, work_mode, logged_by)
		 VALUES (?, ?, ?, ?, 'UTC', 'remote', ?)`,
		presenceOwnerID, presenceBusinessID, onsiteEmployeeID, presenceDate, presenceOnsiteUserID,
	); err != nil {
		t.Fatalf("insert flow_work_days: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM flow_work_days WHERE employee_id = ?", onsiteEmployeeID)
	})

	if _, err := db.ExecContext(ctx,
		`INSERT INTO hr_leave_types (owner_id, business_id, name, is_paid, requires_approval, is_active)
		 VALUES (?, ?, 'Presence Test Leave', true, false, true)`,
		presenceOwnerID, presenceBusinessID,
	); err != nil {
		t.Fatalf("insert hr_leave_types: %v", err)
	}
	var leaveTypeID int64
	if err := db.QueryRowContext(ctx,
		`SELECT id FROM hr_leave_types WHERE business_id = ? AND name = 'Presence Test Leave'`, presenceBusinessID,
	).Scan(&leaveTypeID); err != nil {
		t.Fatalf("read back hr_leave_types id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_leave_types WHERE id = ?", leaveTypeID)
	})

	// hr_leave scopes owner_id/business_id to the raw caller id, not the
	// tenant's resolved businessID (see the ListPresenceForBusiness doc
	// comment) - so this row must be written under presenceCallerUserID,
	// the same id ApprovedLeaveEmployeeIDs is queried with below, or the
	// lookup silently finds nothing.
	if _, err := db.ExecContext(ctx,
		`INSERT INTO hr_leave_requests (owner_id, business_id, employee_id, leave_type_id, start_date, end_date, total_days, status)
		 VALUES (?, ?, ?, ?, ?, ?, 1, 'approved')`,
		presenceCallerUserID, presenceCallerUserID, leaveEmployeeID, leaveTypeID, presenceDate, presenceDate,
	); err != nil {
		t.Fatalf("insert hr_leave_requests: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_leave_requests WHERE employee_id = ?", leaveEmployeeID)
	})

	entries, err := wsModule.ListPresenceForBusiness(ctx, presenceOwnerID, presenceBusinessID, presenceCallerUserID, presenceDate)
	if err != nil {
		t.Fatalf("ListPresenceForBusiness: %v", err)
	}

	byEmployeeID := make(map[int64]workspaces.PresenceEntry, len(entries))
	for _, e := range entries {
		byEmployeeID[e.EmployeeID] = e
	}

	if got := byEmployeeID[onsiteEmployeeID].Status; got != "remote" {
		t.Errorf("onsite employee status = %q, want %q", got, "remote")
	}
	if got := byEmployeeID[leaveEmployeeID].Status; got != workspaces.PresenceOnLeave {
		t.Errorf("on-leave employee status = %q, want %q", got, workspaces.PresenceOnLeave)
	}
	idleEntry, ok := byEmployeeID[callerEmployeeID]
	if !ok {
		t.Fatal("the caller's own employee row should appear in their own business's presence list")
	}
	if idleEntry.Status != workspaces.PresenceNotStarted {
		t.Errorf("caller (no work day, no leave) status = %q, want %q", idleEntry.Status, workspaces.PresenceNotStarted)
	}
}

func TestListPresenceForBusiness_DeniesCallerWithoutViewDirectory(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	wsModule := newPresenceModule(ctx, db)

	unauthorizedUserID := int64(930099)
	insertPresenceEmployee(t, db, unauthorizedUserID, "Unauthorized")

	_, err := wsModule.ListPresenceForBusiness(ctx, presenceOwnerID, presenceBusinessID, unauthorizedUserID, presenceDate)
	if err == nil {
		t.Fatal("want an error for a caller with no view_directory permission")
	}
}

// Distinct tenant for the isolation test below. Its own business is separate
// from the one the other tests seed, so a failure here cannot be another test's
// leftover row.
const (
	isoOwnerID      int64 = 930300
	isoBusinessID   int64 = 930400
	isoCallerUserID int64 = 930301
	isoOwnUserID    int64 = 930302
	isoOtherUserID  int64 = 930303
)

// ListPresenceForBusiness joins hr_employees, flow_work_days and hr_leave across
// module schemas, which is the shape every proposed cross-app join takes. The
// two tests above seed one tenant, so they pass whether or not the business_id
// filter exists. This one seeds two and asserts the second is absent.
//
// Both halves are asserted through testutil.AssertTenantIsolation: absence alone
// would pass if the query returned nothing at all.
func TestListPresenceForBusiness_DoesNotLeakAnotherTenantsEmployees(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	wsModule := newPresenceModule(ctx, db)

	callerEmployeeID := insertPresenceEmployeeFor(t, db, isoOwnerID, isoBusinessID, isoCallerUserID, "IsoCaller")
	ownEmployeeID := insertPresenceEmployeeFor(t, db, isoOwnerID, isoBusinessID, isoOwnUserID, "IsoColleague")

	// The other tenant. Same table, same date, different business.
	otherEmployeeID := insertPresenceEmployeeFor(
		t, db, testutil.OtherOwnerID, testutil.OtherBusinessID, isoOtherUserID, "IsoOutsider",
	)

	// The caller needs view_directory, granted the same way the test above does
	// it: a real role row plus a real assignment.
	if _, err := db.ExecContext(ctx,
		`INSERT INTO hr_roles (owner_id, business_id, name, description, permissions, is_system)
		 VALUES (?, ?, 'Iso Test Role', '', '{"view_directory": true}', false)`,
		isoOwnerID, isoBusinessID,
	); err != nil {
		t.Fatalf("insert hr_roles: %v", err)
	}
	var roleID int64
	if err := db.QueryRowContext(ctx,
		`SELECT id FROM hr_roles WHERE business_id = ? AND name = 'Iso Test Role'`, isoBusinessID,
	).Scan(&roleID); err != nil {
		t.Fatalf("read back hr_roles id: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), "DELETE FROM hr_roles WHERE id = ?", roleID) })
	if _, err := db.ExecContext(ctx,
		`INSERT INTO hr_employee_roles (owner_id, business_id, employee_id, role_id) VALUES (?, ?, ?, ?)`,
		isoOwnerID, isoBusinessID, callerEmployeeID, roleID,
	); err != nil {
		t.Fatalf("insert hr_employee_roles: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_employee_roles WHERE role_id = ?", roleID)
	})

	// A work day for the other tenant's employee, so the row is reachable by any
	// query that forgets to scope: the leak would be a real presence entry, not
	// an employee with nothing to report.
	if _, err := db.ExecContext(ctx,
		`INSERT INTO flow_work_days (owner_id, business_id, employee_id, date, timezone, work_mode, logged_by)
		 VALUES (?, ?, ?, ?, 'UTC', 'remote', ?)`,
		testutil.OtherOwnerID, testutil.OtherBusinessID, otherEmployeeID, presenceDate, isoOtherUserID,
	); err != nil {
		t.Fatalf("insert flow_work_days for the other tenant: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM flow_work_days WHERE employee_id = ?", otherEmployeeID)
	})

	entries, err := wsModule.ListPresenceForBusiness(ctx, isoOwnerID, isoBusinessID, isoCallerUserID, presenceDate)
	if err != nil {
		t.Fatalf("ListPresenceForBusiness: %v", err)
	}

	got := make([]int64, 0, len(entries))
	for _, e := range entries {
		got = append(got, e.EmployeeID)
	}

	testutil.AssertTenantIsolation(t,
		got,
		[]int64{callerEmployeeID, ownEmployeeID},
		[]int64{otherEmployeeID},
	)
}
