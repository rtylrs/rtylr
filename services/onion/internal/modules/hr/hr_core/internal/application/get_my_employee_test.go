package application

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/scopes"

	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// TestGetEmployeeIDByUserID_ThenGetEmployeeForViewer_ResolvesOwnRecord proves
// the composition GetMyEmployeeHandler relies on: resolving a caller's own
// employeeID and then fetching it returns their own record, compensation
// included, with no separate self-lookup RPC needed anywhere in the contract.
func TestGetEmployeeIDByUserID_ThenGetEmployeeForViewer_ResolvesOwnRecord(t *testing.T) {
	db := testutil.OpenDB(t)
	svc := &Service{repo: infrastructure.NewRepository(db).WithAccountEmailReader(scopes.NewReader(db).AccountEmail)}
	ctx := context.Background()

	const linkedUserID int64 = 900101
	created, err := svc.CreateEmployee(ctx, testutil.OwnerID, domain.EmployeeRecord{
		UserID:         linkedUserID,
		FirstName:      "Self",
		LastName:       "Lookup",
		Email:          "self.lookup@example.com",
		Status:         "active",
		EmploymentType: "full_time",
		SalaryCents:    600000,
	})
	if err != nil {
		t.Fatalf("CreateEmployee: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_employees WHERE id = ?", created.ID)
	})

	employeeID, err := svc.GetEmployeeIDByUserID(ctx, testutil.OwnerID, linkedUserID)
	if err != nil {
		t.Fatalf("GetEmployeeIDByUserID: %v", err)
	}
	if employeeID != created.ID {
		t.Fatalf("employeeID = %d, want %d", employeeID, created.ID)
	}

	emp, err := svc.GetEmployeeForViewer(ctx, testutil.OwnerID, linkedUserID, employeeID)
	if err != nil {
		t.Fatalf("GetEmployeeForViewer: %v", err)
	}
	if emp.ID != created.ID || emp.FirstName != "Self" {
		t.Fatalf("resolved employee = %+v, want the caller's own record", emp)
	}
	if emp.SalaryCents != 600000 {
		t.Fatalf("SalaryCents = %v, want 600000 - a caller must see their own compensation", emp.SalaryCents)
	}
}
