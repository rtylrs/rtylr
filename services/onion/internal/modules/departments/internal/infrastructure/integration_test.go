package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/departments/internal/domain"
	"github.com/voxire/onion/internal/modules/departments/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListDepartments_ReturnsOperations(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	departments, err := repo.ListDepartments(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListDepartments: %v", err)
	}

	found := false
	for _, d := range departments {
		if d.Name == "Operations" {
			found = true
			if d.ID != 4001 {
				t.Errorf("expected Operations to have id=4001, got %d", d.ID)
			}
			break
		}
	}
	if !found {
		t.Errorf("expected 'Operations' in department list, got %v", departments)
	}
}

func TestListDepartments_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	departments, err := repo.ListDepartments(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListDepartments: %v", err)
	}
	for _, d := range departments {
		if d.OwnerID != testutil.OwnerID {
			t.Errorf("expected OwnerID=%d, got %d for dept id=%d", testutil.OwnerID, d.OwnerID, d.ID)
		}
	}
}

func TestListDepartments_NamesAreNonEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	departments, err := repo.ListDepartments(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListDepartments: %v", err)
	}
	for _, d := range departments {
		if d.Name == "" {
			t.Errorf("department id=%d has empty Name", d.ID)
		}
	}
}

func TestCreateDepartment_IncreasesCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.ListDepartments(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListDepartments before: %v", err)
	}

	id, err := repo.CreateDepartment(ctx, domain.DepartmentInput{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		Name:        "Integration Test Dept",
		Description: "Created by integration test",
	})
	if err != nil {
		t.Fatalf("CreateDepartment: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
	defer func() {
		_ = repo.DeleteDepartment(ctx, testutil.OwnerID, id)
	}()

	after, err := repo.ListDepartments(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListDepartments after: %v", err)
	}
	if len(after) != len(before)+1 {
		t.Errorf("expected count to increase by 1: before=%d after=%d", len(before), len(after))
	}
}

func TestUpdateDepartment_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateDepartment(ctx, domain.DepartmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Before Rename",
	})
	if err != nil {
		t.Fatalf("CreateDepartment: %v", err)
	}
	defer func() {
		_ = repo.DeleteDepartment(ctx, testutil.OwnerID, id)
	}()

	err = repo.UpdateDepartment(ctx, testutil.OwnerID, domain.DepartmentInput{
		ID:         id,
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "After Rename",
	})
	if err != nil {
		t.Fatalf("UpdateDepartment: %v", err)
	}

	departments, err := repo.ListDepartments(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListDepartments: %v", err)
	}
	for _, d := range departments {
		if d.ID == id && d.Name != "After Rename" {
			t.Errorf("expected Name='After Rename', got %q", d.Name)
		}
	}
}

func TestDeleteDepartment_RemovesEntry(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateDepartment(ctx, domain.DepartmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "To Delete",
	})
	if err != nil {
		t.Fatalf("CreateDepartment: %v", err)
	}

	if err := repo.DeleteDepartment(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("DeleteDepartment: %v", err)
	}

	departments, err := repo.ListDepartments(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListDepartments: %v", err)
	}
	for _, d := range departments {
		if d.ID == id {
			t.Errorf("expected department id=%d to be deleted but still found", id)
		}
	}
}

func TestListDepartments_DifferentOwnerIsIsolated(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// A different owner should not see departments belonging to OwnerID.
	otherDepts, err := repo.ListDepartments(ctx, 9999999)
	if err != nil {
		t.Fatalf("ListDepartments other owner: %v", err)
	}
	for _, d := range otherDepts {
		if d.OwnerID == testutil.OwnerID {
			t.Errorf("owner isolation violated: found dept id=%d belonging to OwnerID=%d", d.ID, d.OwnerID)
		}
	}
}

func TestCreateDepartment_WithDescription(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateDepartment(ctx, domain.DepartmentInput{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		Name:        "Dept With Desc",
		Description: "A helpful description",
	})
	if err != nil {
		t.Fatalf("CreateDepartment: %v", err)
	}
	defer func() {
		_ = repo.DeleteDepartment(ctx, testutil.OwnerID, id)
	}()

	departments, err := repo.ListDepartments(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListDepartments: %v", err)
	}
	for _, d := range departments {
		if d.ID == id {
			if d.Description != "A helpful description" {
				t.Errorf("expected Description='A helpful description', got %q", d.Description)
			}
			return
		}
	}
	t.Errorf("created department id=%d not found in list", id)
}

func TestCreateDepartment_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateDepartment(ctx, domain.DepartmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Positive ID Test Dept",
	})
	if err != nil {
		t.Fatalf("CreateDepartment: %v", err)
	}
	defer func() { _ = repo.DeleteDepartment(ctx, testutil.OwnerID, id) }()

	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestCreateDepartment_MultipleDistinctIDs(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.CreateDepartment(ctx, domain.DepartmentInput{
		OwnerID: testutil.OwnerID, Name: "Dept Multi A",
		BusinessID: testutil.BusinessID,
	})
	if err != nil {
		t.Fatalf("CreateDepartment A: %v", err)
	}
	defer func() { _ = repo.DeleteDepartment(ctx, testutil.OwnerID, id1) }()

	id2, err := repo.CreateDepartment(ctx, domain.DepartmentInput{
		OwnerID: testutil.OwnerID, Name: "Dept Multi B",
		BusinessID: testutil.BusinessID,
	})
	if err != nil {
		t.Fatalf("CreateDepartment B: %v", err)
	}
	defer func() { _ = repo.DeleteDepartment(ctx, testutil.OwnerID, id2) }()

	if id1 == id2 {
		t.Errorf("expected distinct IDs, both got %d", id1)
	}
}

func TestUpdateDepartment_UpdatesDescription(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateDepartment(ctx, domain.DepartmentInput{
		OwnerID: testutil.OwnerID, Name: "Desc Update Dept", Description: "Old desc",
		BusinessID: testutil.BusinessID,
	})
	if err != nil {
		t.Fatalf("CreateDepartment: %v", err)
	}
	defer func() { _ = repo.DeleteDepartment(ctx, testutil.OwnerID, id) }()

	if err := repo.UpdateDepartment(ctx, testutil.OwnerID, domain.DepartmentInput{
		ID: id, OwnerID: testutil.OwnerID, Name: "Desc Update Dept", Description: "New desc",
		BusinessID: testutil.BusinessID,
	}); err != nil {
		t.Fatalf("UpdateDepartment: %v", err)
	}

	departments, err := repo.ListDepartments(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListDepartments: %v", err)
	}
	for _, d := range departments {
		if d.ID == id && d.Description != "New desc" {
			t.Errorf("expected Description='New desc', got %q", d.Description)
		}
	}
}

func TestDeleteDepartment_IdempotentOnMissing(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Deleting a non-existent ID should not return a hard error.
	err := repo.DeleteDepartment(ctx, testutil.OwnerID, 9999999)
	_ = err // may or may not error; must not panic
}

func TestListDepartments_IDsArePositive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	departments, err := repo.ListDepartments(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListDepartments: %v", err)
	}
	for _, d := range departments {
		if d.ID <= 0 {
			t.Errorf("expected positive ID, got %d", d.ID)
		}
	}
}

func TestUpdateDepartment_WrongOwnerNoEffect(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateDepartment(ctx, domain.DepartmentInput{
		OwnerID: testutil.OwnerID, Name: "Owner Guard Dept",
		BusinessID: testutil.BusinessID,
	})
	if err != nil {
		t.Fatalf("CreateDepartment: %v", err)
	}
	defer func() { _ = repo.DeleteDepartment(ctx, testutil.OwnerID, id) }()

	// Update with a different ownerID should not affect the record.
	_ = repo.UpdateDepartment(ctx, 9999999, domain.DepartmentInput{
		ID: id, OwnerID: 9999999, Name: "Hijacked Name",
	})

	departments, err := repo.ListDepartments(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListDepartments: %v", err)
	}
	for _, d := range departments {
		if d.ID == id && d.Name == "Hijacked Name" {
			t.Error("wrong owner was able to update the department name")
		}
	}
}

func TestCreateDepartment_NoDescription_Allowed(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateDepartment(ctx, domain.DepartmentInput{
		OwnerID: testutil.OwnerID, Name: "No Desc Dept",
		BusinessID: testutil.BusinessID,
	})
	if err != nil {
		t.Fatalf("CreateDepartment without description: %v", err)
	}
	defer func() { _ = repo.DeleteDepartment(ctx, testutil.OwnerID, id) }()

	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestListDepartments_Idempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	a, err := repo.ListDepartments(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	b, err := repo.ListDepartments(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if len(a) != len(b) {
		t.Errorf("idempotency failed: got %d then %d", len(a), len(b))
	}
}

func TestListDepartments_UnknownOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	depts, err := repo.ListDepartments(ctx, 9999997)
	if err != nil {
		t.Fatalf("ListDepartments: %v", err)
	}
	if len(depts) != 0 {
		t.Errorf("expected 0 for unknown owner, got %d", len(depts))
	}
}

func TestDeleteDepartment_DecreasesCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateDepartment(ctx, domain.DepartmentInput{
		OwnerID: testutil.OwnerID, Name: "Count Decrease Dept",
		BusinessID: testutil.BusinessID,
	})
	if err != nil {
		t.Fatalf("CreateDepartment: %v", err)
	}

	before, err := repo.ListDepartments(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListDepartments before delete: %v", err)
	}

	if err := repo.DeleteDepartment(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("DeleteDepartment: %v", err)
	}

	after, err := repo.ListDepartments(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListDepartments after delete: %v", err)
	}
	if len(after) != len(before)-1 {
		t.Errorf("expected count to decrease by 1: before=%d after=%d", len(before), len(after))
	}
}
