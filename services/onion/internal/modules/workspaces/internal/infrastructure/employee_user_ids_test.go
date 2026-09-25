package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// Work days are keyed on hr_employees.id; every roster the client holds is keyed
// on users.id. EmployeeUserIDs is the join between them, and without it the Team
// and Reports screens cannot attach a day to a person (#1072).
func TestEmployeeUserIDsResolvesOnlyEmployeesWithAUser(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := infrastructure.NewRepository(db)

	// The HR fixtures live under business 1001; discover an employee rather
	// than hardcoding an id, so a reshuffled fixture set fails loudly here
	// instead of quietly testing nothing.
	const fixtureBusinessID int64 = 1001
	var employeeID int64
	if err := db.QueryRowContext(ctx,
		"SELECT id FROM hr_employees WHERE business_id = ? ORDER BY id ASC LIMIT 1",
		fixtureBusinessID,
	).Scan(&employeeID); err != nil {
		t.Skipf("no hr_employees fixture for business %d: %v", fixtureBusinessID, err)
	}

	var original *int64
	if err := db.QueryRowContext(ctx,
		"SELECT user_id FROM hr_employees WHERE id = ?", employeeID,
	).Scan(&original); err != nil {
		t.Fatalf("read original user_id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "UPDATE hr_employees SET user_id = ? WHERE id = ?", original, employeeID)
	})

	// An employee with no user has no account to attribute a day to, so it is
	// absent from the map rather than present with a zero.
	if _, err := db.ExecContext(ctx, "UPDATE hr_employees SET user_id = NULL WHERE id = ?", employeeID); err != nil {
		t.Fatalf("clear user_id: %v", err)
	}
	byEmployeeID, err := repo.EmployeeUserIDs(ctx, fixtureBusinessID, []int64{employeeID})
	if err != nil {
		t.Fatalf("EmployeeUserIDs with a null user: %v", err)
	}
	if _, present := byEmployeeID[employeeID]; present {
		t.Errorf("employee %d has no user_id but was resolved to %d", employeeID, byEmployeeID[employeeID])
	}

	// With a user, the day can be attributed.
	const sentinelUserID int64 = 987654
	if _, err := db.ExecContext(ctx, "UPDATE hr_employees SET user_id = ? WHERE id = ?", sentinelUserID, employeeID); err != nil {
		t.Fatalf("set user_id: %v", err)
	}
	byEmployeeID, err = repo.EmployeeUserIDs(ctx, fixtureBusinessID, []int64{employeeID})
	if err != nil {
		t.Fatalf("EmployeeUserIDs: %v", err)
	}
	if byEmployeeID[employeeID] != sentinelUserID {
		t.Errorf("employee %d resolved to %d, want %d", employeeID, byEmployeeID[employeeID], sentinelUserID)
	}

	// Another business must not resolve the same employee.
	crossTenant, err := repo.EmployeeUserIDs(ctx, fixtureBusinessID+1, []int64{employeeID})
	if err != nil {
		t.Fatalf("EmployeeUserIDs across tenants: %v", err)
	}
	if _, present := crossTenant[employeeID]; present {
		t.Errorf("employee %d of business %d resolved under a different business", employeeID, fixtureBusinessID)
	}
}

func TestEmployeeUserIDsHandlesAnEmptyRequest(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	byEmployeeID, err := repo.EmployeeUserIDs(context.Background(), 1001, nil)
	if err != nil {
		t.Fatalf("EmployeeUserIDs with no ids: %v", err)
	}
	if len(byEmployeeID) != 0 {
		t.Errorf("resolved %d entries for an empty request", len(byEmployeeID))
	}
}
