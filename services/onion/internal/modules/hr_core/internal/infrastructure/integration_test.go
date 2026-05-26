package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_core/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListEmployees_ReturnsAliceAndBob(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	employees, total, err := repo.ListEmployees(ctx, testutil.OwnerID, domain.ListEmployeesParams{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListEmployees: %v", err)
	}
	if total < 2 {
		t.Fatalf("expected at least 2 employees, got total=%d", total)
	}

	names := make(map[string]bool)
	for _, e := range employees {
		names[e.FirstName] = true
	}
	if !names["Alice"] {
		t.Error("expected Alice in employee list")
	}
	if !names["Bob"] {
		t.Error("expected Bob in employee list")
	}
}

func TestGetEmployee_AliceHasSalary1200(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	emp, err := repo.GetEmployee(ctx, testutil.OwnerID, 4101)
	if err != nil {
		t.Fatalf("GetEmployee: %v", err)
	}
	if emp.FirstName != "Alice" {
		t.Errorf("expected FirstName=Alice, got %q", emp.FirstName)
	}
	if emp.Salary != 1200.0 {
		t.Errorf("expected Salary=1200, got %v", emp.Salary)
	}
}

func TestListDepartments_ReturnsOperations(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	departments, err := repo.ListDepartments(ctx, testutil.OwnerID, 0)
	if err != nil {
		t.Fatalf("ListDepartments: %v", err)
	}

	found := false
	for _, d := range departments {
		if d.Name == "Operations" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected department 'Operations' in list, got %v", departments)
	}
}
