package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
	"github.com/voxire/onion/internal/modules/hr/staff/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestEnsureStaffForUserMirrorsTheHRCoreEmployee(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	core := hrcore.NewModule(ctx, db)
	repo := infrastructure.NewRepository(db)
	userID := time.Now().UnixNano()
	email := fmt.Sprintf("hr-staff-owner-%d@example.test", userID)

	if err := core.EnsureEmployeeForUser(ctx, testutil.BusinessID, userID, "Ada", "Lovelace", email); err != nil {
		t.Fatalf("ensure HR employee: %v", err)
	}
	t.Cleanup(func() {
		if err := repo.DeleteStaffForUser(context.Background(), testutil.BusinessID, userID); err != nil {
			t.Errorf("delete staff record: %v", err)
		}
		if err := core.DeleteEmployeeForUser(context.Background(), testutil.BusinessID, userID); err != nil {
			t.Errorf("delete employee record: %v", err)
		}
	})

	if err := repo.EnsureStaffForUser(ctx, testutil.BusinessID, userID, "member"); err != nil {
		t.Fatalf("ensure staff: %v", err)
	}
	if err := repo.EnsureStaffForUser(ctx, testutil.BusinessID, userID, "changed"); err != nil {
		t.Fatalf("ensure staff a second time: %v", err)
	}

	var staffCount int
	var firstName, lastName, gotEmail, role string
	var employeeID int64
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*), MAX(first_name), MAX(last_name), MAX(email), MAX(role), MAX(employee_id)
		FROM hr_staff WHERE owner_id = ? AND business_id = ? AND user_id = ?`,
		testutil.OwnerID, testutil.BusinessID, userID).Scan(&staffCount, &firstName, &lastName, &gotEmail, &role, &employeeID); err != nil {
		t.Fatalf("read staff record: %v", err)
	}
	if staffCount != 1 || firstName != "Ada" || lastName != "Lovelace" || gotEmail != email || role != "member" || employeeID == 0 {
		t.Fatalf("staff count=%d name=%q %q email=%q role=%q employee=%d", staffCount, firstName, lastName, gotEmail, role, employeeID)
	}
}
