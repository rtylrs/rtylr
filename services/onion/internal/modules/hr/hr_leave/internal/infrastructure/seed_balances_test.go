package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestSeedMissingLeaveBalances_FillsGapsOnce(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const (
		seedYear           int32 = 2099
		activeEmployee     int64 = 4102
		terminatedEmployee int64 = 4190
		typeName                 = "Seed Missing Balances Leave"
	)
	t.Cleanup(func() {
		cleanup := context.Background()
		db.ExecContext(cleanup, "DELETE FROM hr_leave_balances WHERE business_id = ? AND year = ?", testutil.BusinessID, seedYear)
		db.ExecContext(cleanup, "DELETE FROM hr_leave_types WHERE business_id = ? AND name = ?", testutil.BusinessID, typeName)
		db.ExecContext(cleanup, "DELETE FROM hr_employees WHERE id = ?", terminatedEmployee)
	})
	if _, err := db.ExecContext(ctx,
		`INSERT INTO hr_employees (id, owner_id, business_id, first_name, last_name, email, status)
		 VALUES (?, ?, ?, 'Former', 'Employee', 'seed-terminated@test.local', 'terminated')`,
		terminatedEmployee, testutil.BusinessID, testutil.BusinessID); err != nil {
		t.Fatalf("insert terminated employee: %v", err)
	}
	lt, err := repo.CreateLeaveType(ctx, testutil.BusinessID, domain.LeaveTypeRecord{Name: typeName, IsActive: true})
	if err != nil {
		t.Fatalf("CreateLeaveType: %v", err)
	}

	countFor := func(employeeID int64) int {
		t.Helper()
		var rows int
		if err := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM hr_leave_balances WHERE business_id = ? AND employee_id = ? AND leave_type_id = ? AND year = ?",
			testutil.BusinessID, employeeID, lt.ID, seedYear).Scan(&rows); err != nil {
			t.Fatalf("count balances: %v", err)
		}
		return rows
	}

	seeded, err := repo.SeedMissingLeaveBalances(ctx, testutil.BusinessID, activeEmployee, seedYear)
	if err != nil {
		t.Fatalf("SeedMissingLeaveBalances for one employee: %v", err)
	}
	if seeded == 0 || countFor(activeEmployee) != 1 {
		t.Fatalf("seeded = %d, balance rows = %d, want the missing row created", seeded, countFor(activeEmployee))
	}
	if countFor(4101) != 0 {
		t.Fatal("an employee-scoped seed created a balance for another employee")
	}

	if _, err := repo.SeedMissingLeaveBalances(ctx, testutil.BusinessID, 0, seedYear); err != nil {
		t.Fatalf("SeedMissingLeaveBalances for the business: %v", err)
	}
	if countFor(4101) != 1 || countFor(activeEmployee) != 1 {
		t.Fatalf("business seed left rows = %d and %d, want exactly one each", countFor(4101), countFor(activeEmployee))
	}
	if countFor(terminatedEmployee) != 0 {
		t.Fatal("a terminated employee was given a leave balance")
	}

	again, err := repo.SeedMissingLeaveBalances(ctx, testutil.BusinessID, 0, seedYear)
	if err != nil {
		t.Fatalf("repeat SeedMissingLeaveBalances: %v", err)
	}
	if again != 0 {
		t.Fatalf("repeat seed inserted %d rows, want 0", again)
	}
}
