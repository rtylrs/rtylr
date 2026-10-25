package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestEnsureEmployeeForUserOwnsProvisioningAndTerminationReads(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := infrastructure.NewRepository(db)
	userID := time.Now().UnixNano()
	email := fmt.Sprintf("hr-owner-%d@example.test", userID)

	t.Cleanup(func() {
		if err := repo.DeleteEmployeeForUser(context.Background(), testutil.BusinessID, userID); err != nil {
			t.Errorf("delete employee: %v", err)
		}
	})

	if err := repo.EnsureEmployeeForUser(ctx, testutil.BusinessID, userID, "Ada", "Lovelace", email); err != nil {
		t.Fatalf("ensure employee: %v", err)
	}
	if err := repo.EnsureEmployeeForUser(ctx, testutil.BusinessID, userID, "Grace", "Hopper", "changed@example.test"); err != nil {
		t.Fatalf("ensure employee a second time: %v", err)
	}

	var businessID int64
	var firstName, lastName, gotEmail string
	if err := db.QueryRowContext(ctx, `
		SELECT business_id, first_name, last_name, email
		FROM hr_employees WHERE user_id = ?`, userID).Scan(&businessID, &firstName, &lastName, &gotEmail); err != nil {
		t.Fatalf("read employee: %v", err)
	}
	if businessID != testutil.BusinessID || firstName != "Ada" || lastName != "Lovelace" || gotEmail != email {
		t.Fatalf("employee = business %d, %q %q <%s>", businessID, firstName, lastName, gotEmail)
	}

	if _, err := db.ExecContext(ctx,
		"UPDATE hr_employees SET status = 'terminated' WHERE user_id = ? AND business_id = ?", userID, testutil.BusinessID); err != nil {
		t.Fatalf("terminate employee: %v", err)
	}
	terminated, err := repo.ListTerminatedEmployees(ctx)
	if err != nil {
		t.Fatalf("list terminated employees: %v", err)
	}
	found := false
	for _, employee := range terminated {
		if employee.BusinessID == testutil.BusinessID && employee.UserID == userID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("terminated employee %d/%d not returned", testutil.BusinessID, userID)
	}
}
