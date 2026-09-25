package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_core/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListEmployees_TotalIsNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, total, err := repo.ListEmployees(ctx, testutil.OwnerID, domain.ListEmployeesParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListEmployees: %v", err)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %d", total)
	}
}

func TestListEmployees_LimitRespected(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	employees, _, err := repo.ListEmployees(ctx, testutil.OwnerID, domain.ListEmployeesParams{Limit: 1, Offset: 0})
	if err != nil {
		t.Fatalf("ListEmployees: %v", err)
	}
	if len(employees) > 1 {
		t.Errorf("expected at most 1 employee with limit=1, got %d", len(employees))
	}
}

func TestListEmployees_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	employees, total, err := repo.ListEmployees(ctx, 0, domain.ListEmployeesParams{Limit: 50, Offset: 0})
	if err != nil {
		t.Fatalf("ListEmployees zero owner: %v", err)
	}
	if total != 0 || len(employees) != 0 {
		t.Errorf("expected empty for owner 0, got total=%d records=%d", total, len(employees))
	}
}

func TestListEmployees_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	employees, _, err := repo.ListEmployees(ctx, testutil.OwnerID, domain.ListEmployeesParams{Limit: 100, Offset: 0})
	if err != nil {
		t.Fatalf("ListEmployees: %v", err)
	}
	for _, e := range employees {
		if e.OwnerID != testutil.OwnerID {
			t.Errorf("employee %d has owner_id %d, want %d", e.ID, e.OwnerID, testutil.OwnerID)
		}
	}
}

func TestListEmployees_PaginationOffsetChangesResults(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, total, err := repo.ListEmployees(ctx, testutil.OwnerID, domain.ListEmployeesParams{Limit: 1, Offset: 0})
	if err != nil {
		t.Fatalf("ListEmployees page 1: %v", err)
	}
	if total < 2 {
		t.Skip("not enough employees to test pagination")
	}
	second, _, err := repo.ListEmployees(ctx, testutil.OwnerID, domain.ListEmployeesParams{Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("ListEmployees page 2: %v", err)
	}
	if len(first) > 0 && len(second) > 0 && first[0].ID == second[0].ID {
		t.Error("pagination offset had no effect")
	}
}

func TestGetEmployee_Alice_HasCorrectFields(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	emp, err := repo.GetEmployee(ctx, testutil.BusinessID, 4101)
	if err != nil {
		t.Fatalf("GetEmployee: %v", err)
	}
	if emp.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", emp.OwnerID, testutil.OwnerID)
	}
	if emp.Email == "" {
		t.Error("expected non-empty email for Alice")
	}
}

func TestGetEmployee_UnknownID_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetEmployee(ctx, testutil.OwnerID, 999999999)
	if err == nil {
		t.Error("expected error for unknown employee, got nil")
	}
}

func TestCreateEmployee_PersistsName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	emp, err := repo.CreateEmployee(ctx, testutil.OwnerID, domain.EmployeeRecord{
		FirstName:      "Test",
		LastName:       "Worker",
		Email:          "test.worker.infra@example.com",
		Status:         "active",
		EmploymentType: "full_time",
	})
	if err != nil {
		t.Fatalf("CreateEmployee: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_employees WHERE id = ?", emp.ID)
	})
	if emp.FirstName != "Test" {
		t.Errorf("FirstName = %q, want 'Test'", emp.FirstName)
	}
	if emp.ID == 0 {
		t.Error("expected non-zero ID after create")
	}
}

func TestCreateEmployee_OwnerIDSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	emp, err := repo.CreateEmployee(ctx, testutil.OwnerID, domain.EmployeeRecord{
		FirstName:      "Owner",
		LastName:       "Check",
		Email:          "owner.check.infra@example.com",
		Status:         "active",
		EmploymentType: "full_time",
	})
	if err != nil {
		t.Fatalf("CreateEmployee: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_employees WHERE id = ?", emp.ID)
	})
	if emp.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", emp.OwnerID, testutil.OwnerID)
	}
}

func TestGetEmployeeIDByUserID_ReturnsMatchingEmployee(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const linkedUserID int64 = 900001
	created, err := repo.CreateEmployee(ctx, testutil.OwnerID, domain.EmployeeRecord{
		UserID:         linkedUserID,
		FirstName:      "Linked",
		LastName:       "User",
		Email:          "linked.user.infra@example.com",
		Status:         "active",
		EmploymentType: "full_time",
	})
	if err != nil {
		t.Fatalf("CreateEmployee: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_employees WHERE id = ?", created.ID)
	})

	employeeID, err := repo.GetEmployeeIDByUserID(ctx, testutil.OwnerID, linkedUserID)
	if err != nil {
		t.Fatalf("GetEmployeeIDByUserID: %v", err)
	}
	if employeeID != created.ID {
		t.Errorf("employeeID = %d, want %d", employeeID, created.ID)
	}
}

func TestGetEmployeeIDByUserID_UnknownUser_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetEmployeeIDByUserID(ctx, testutil.OwnerID, 999999999)
	if !errors.Is(err, domain.ErrEmployeeNotFound) {
		t.Errorf("expected ErrEmployeeNotFound, got %v", err)
	}
}

func TestUpdateEmployee_StatusChanges(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateEmployee(ctx, testutil.OwnerID, domain.EmployeeRecord{
		FirstName:      "Status",
		LastName:       "Change",
		Email:          "status.change.infra@example.com",
		Status:         "active",
		EmploymentType: "full_time",
	})
	if err != nil {
		t.Fatalf("CreateEmployee: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_employees WHERE id = ?", created.ID)
	})

	created.Status = "inactive"
	updated, err := repo.UpdateEmployee(ctx, testutil.OwnerID, *created)
	if err != nil {
		t.Fatalf("UpdateEmployee: %v", err)
	}
	if updated.Status != "inactive" {
		t.Errorf("Status = %q, want 'inactive'", updated.Status)
	}
}
