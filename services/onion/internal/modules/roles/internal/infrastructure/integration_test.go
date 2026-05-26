package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/roles/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListRoles_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	roles, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("roles.List: unexpected error: %v", err)
	}
	if roles == nil {
		t.Fatal("roles.List: expected non-nil slice, got nil")
	}
}

func TestListRoles_WrongOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	roles, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("roles.List(ownerID=0): unexpected error: %v", err)
	}
	if len(roles) != 0 {
		t.Fatalf("roles.List(ownerID=0): expected 0 rows, got %d", len(roles))
	}
}

func TestCreateRole_ReturnsInsertedID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, "Test Role", []string{"read", "write"})
	if err != nil {
		t.Fatalf("roles.Create: unexpected error: %v", err)
	}
	if id <= 0 {
		t.Fatalf("roles.Create: expected positive ID, got %d", id)
	}
}

func TestCreateRole_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, "Appears In List Role", []string{"read"})
	if err != nil {
		t.Fatalf("roles.Create: %v", err)
	}

	roles, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("roles.List: %v", err)
	}

	var found bool
	for _, r := range roles {
		if r.ID == id {
			found = true
			if r.Name != "Appears In List Role" {
				t.Errorf("Name mismatch: got %q", r.Name)
			}
			break
		}
	}
	if !found {
		t.Fatalf("created role id=%d not found in list", id)
	}
}

func TestCreateRole_PermissionsPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	perms := []string{"read", "write", "delete"}

	id, err := repo.Create(ctx, testutil.OwnerID, "Perms Role", perms)
	if err != nil {
		t.Fatalf("roles.Create: %v", err)
	}

	roles, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("roles.List: %v", err)
	}

	for _, r := range roles {
		if r.ID == id {
			if len(r.Permissions) != len(perms) {
				t.Errorf("Permissions length mismatch: got %d, want %d", len(r.Permissions), len(perms))
			}
			return
		}
	}
	t.Fatalf("created role id=%d not found in list", id)
}

func TestUpdateRole_NameChanges(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, "Before Update Role", []string{"read"})
	if err != nil {
		t.Fatalf("roles.Create: %v", err)
	}

	err = repo.Update(ctx, testutil.OwnerID, id, "After Update Role", []string{"write"})
	if err != nil {
		t.Fatalf("roles.Update: %v", err)
	}

	roles, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("roles.List after update: %v", err)
	}

	for _, r := range roles {
		if r.ID == id {
			if r.Name != "After Update Role" {
				t.Errorf("Name not updated: got %q", r.Name)
			}
			return
		}
	}
	t.Fatalf("updated role id=%d not found in list", id)
}

func TestDeleteRole_RemovedFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, "Delete Me Role", []string{})
	if err != nil {
		t.Fatalf("roles.Create: %v", err)
	}

	err = repo.Delete(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("roles.Delete: %v", err)
	}

	roles, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("roles.List after delete: %v", err)
	}

	for _, r := range roles {
		if r.ID == id {
			t.Fatalf("deleted role id=%d still present in list", id)
		}
	}
}

func TestListRoles_EachHasPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, testutil.OwnerID, "ID Check Role", []string{})
	if err != nil {
		t.Fatalf("seed role: %v", err)
	}

	roles, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("roles.List: %v", err)
	}

	for _, r := range roles {
		if r.ID <= 0 {
			t.Errorf("role has non-positive ID: %d", r.ID)
		}
	}
}

func TestListRoles_EachHasNonEmptyName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, testutil.OwnerID, "Name Check Role", []string{})
	if err != nil {
		t.Fatalf("seed role: %v", err)
	}

	roles, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("roles.List: %v", err)
	}

	for _, r := range roles {
		if r.Name == "" {
			t.Errorf("role id=%d has empty name", r.ID)
		}
	}
}

func TestListRoles_PermissionsDeserialised(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, testutil.OwnerID, "Perms Deserialise Role", []string{"manage"})
	if err != nil {
		t.Fatalf("seed role: %v", err)
	}

	roles, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("roles.List: %v", err)
	}

	// All permissions slices should be non-nil (deserialised from JSON)
	for _, r := range roles {
		if r.Permissions == nil {
			t.Errorf("role id=%d has nil Permissions slice", r.ID)
		}
	}
}

func TestListRoles_ScopedToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, testutil.OwnerID, "Scoped Role", []string{})
	if err != nil {
		t.Fatalf("seed role: %v", err)
	}

	other, err := repo.List(ctx, testutil.OwnerID+9999)
	if err != nil {
		t.Fatalf("roles.List(other owner): %v", err)
	}
	for _, r := range other {
		if r.OwnerID == testutil.OwnerID {
			t.Errorf("other-owner list contains ownerID=%d record", testutil.OwnerID)
		}
	}
}
