package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_core/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateRole_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	role, err := repo.CreateRole(ctx, testutil.OwnerID, domain.RoleRecord{
		Name:        "Test Role Infra",
		Description: "A test role",
		Permissions: `{}`,
	})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	if role.ID == 0 {
		t.Error("expected non-zero ID after create")
	}
}

func TestCreateRole_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	role, err := repo.CreateRole(ctx, testutil.OwnerID, domain.RoleRecord{
		Name:        "Name Store Role",
		Permissions: `{}`,
	})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	if role.Name != "Name Store Role" {
		t.Errorf("Name = %q, want 'Name Store Role'", role.Name)
	}
}

func TestCreateRole_OwnerIDSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	role, err := repo.CreateRole(ctx, testutil.OwnerID, domain.RoleRecord{
		Name:        "OwnerID Role",
		Permissions: `{}`,
	})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	if role.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", role.OwnerID, testutil.OwnerID)
	}
}

func TestListRoles_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.ListRoles(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListRoles before: %v", err)
	}

	_, err = repo.CreateRole(ctx, testutil.OwnerID, domain.RoleRecord{
		Name:        "Count Increase Role",
		Permissions: `{}`,
	})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}

	after, err := repo.ListRoles(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListRoles after: %v", err)
	}
	if len(after) <= len(before) {
		t.Errorf("expected count to increase from %d, got %d", len(before), len(after))
	}
}

func TestUpdateRole_DescriptionChanges(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateRole(ctx, testutil.OwnerID, domain.RoleRecord{
		Name:        "Role To Update",
		Description: "before",
		Permissions: `{}`,
	})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}

	created.Description = "after"
	updated, err := repo.UpdateRole(ctx, testutil.OwnerID, *created)
	if err != nil {
		t.Fatalf("UpdateRole: %v", err)
	}
	if updated.Description != "after" {
		t.Errorf("Description = %q, want 'after'", updated.Description)
	}
}

func TestHasPermission_UnknownEmployeeReturnsFalse(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	has, err := repo.HasPermission(ctx, testutil.OwnerID, 999999999, "manage_employees")
	if err != nil {
		t.Fatalf("HasPermission: %v", err)
	}
	if has {
		t.Error("expected false for unknown employee, got true")
	}
}

func TestAssignRole_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	role, err := repo.CreateRole(ctx, testutil.OwnerID, domain.RoleRecord{
		Name:        "Assign Test Role",
		Permissions: `{}`,
	})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}

	err = repo.AssignRole(ctx, testutil.OwnerID, 4101, role.ID)
	if err != nil {
		t.Fatalf("AssignRole: %v", err)
	}
}
