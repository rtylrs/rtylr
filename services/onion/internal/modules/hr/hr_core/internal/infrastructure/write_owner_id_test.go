package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// The write methods below echo the caller's record back instead of re-reading
// the row, so every field the caller did not set has to be filled in from the
// arguments. owner_id is written by the INSERT and published in the API
// response, so a returned record missing it makes a create or update answer
// disagree with the subsequent read.
//
// The update cases clear OwnerID on the input on purpose: the route handlers
// build a fresh record from the request body and never set it, so echoing the
// input back is exactly how the field goes missing.

func TestCreateEmployeeWithBalances_OwnerIDSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	emp, err := repo.CreateEmployeeWithBalances(ctx, testutil.OwnerID, domain.EmployeeRecord{
		FirstName: "Balances", LastName: "Owner", Email: "balances.owner.infra@example.com",
		Status: "active", EmploymentType: "full_time",
	}, []int64{})
	if err != nil {
		t.Fatalf("CreateEmployeeWithBalances: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_employees WHERE id = ?", emp.ID)
	})

	if emp.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", emp.OwnerID, testutil.OwnerID)
	}
}

func TestUpdateEmployee_OwnerIDSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateEmployee(ctx, testutil.OwnerID, domain.EmployeeRecord{
		FirstName: "Update", LastName: "Owner", Email: "update.owner.infra@example.com",
		Status: "active", EmploymentType: "full_time",
	})
	if err != nil {
		t.Fatalf("CreateEmployee: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_employees WHERE id = ?", created.ID)
	})

	input := *created
	input.OwnerID = 0
	updated, err := repo.UpdateEmployee(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("UpdateEmployee: %v", err)
	}
	if updated.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", updated.OwnerID, testutil.OwnerID)
	}
}

func TestUpdateLocation_OwnerIDSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateLocation(ctx, testutil.OwnerID, domain.LocationRecord{
		Name: "Owner Update Location", Country: "TC",
	})
	if err != nil {
		t.Fatalf("CreateLocation: %v", err)
	}

	input := *created
	input.OwnerID = 0
	updated, err := repo.UpdateLocation(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("UpdateLocation: %v", err)
	}
	if updated.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", updated.OwnerID, testutil.OwnerID)
	}
}

func TestUpdateDepartment_OwnerIDSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateDepartment(ctx, testutil.OwnerID, domain.DepartmentRecord{
		Name: "Owner Update Department",
	})
	if err != nil {
		t.Fatalf("CreateDepartment: %v", err)
	}

	input := *created
	input.OwnerID = 0
	updated, err := repo.UpdateDepartment(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("UpdateDepartment: %v", err)
	}
	if updated.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", updated.OwnerID, testutil.OwnerID)
	}
}

func TestUpdateWorkSchedule_OwnerIDSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateWorkSchedule(ctx, testutil.OwnerID, domain.WorkScheduleRecord{
		Name: "Owner Update Schedule", ScheduleData: `{}`,
	})
	if err != nil {
		t.Fatalf("CreateWorkSchedule: %v", err)
	}

	input := *created
	input.OwnerID = 0
	updated, err := repo.UpdateWorkSchedule(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("UpdateWorkSchedule: %v", err)
	}
	if updated.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", updated.OwnerID, testutil.OwnerID)
	}
}

func TestUpdateRole_OwnerIDSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateRole(ctx, testutil.OwnerID, domain.RoleRecord{
		Name: "Owner Update Role", Permissions: `{}`,
	})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}

	input := *created
	input.OwnerID = 0
	updated, err := repo.UpdateRole(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("UpdateRole: %v", err)
	}
	if updated.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", updated.OwnerID, testutil.OwnerID)
	}
}
