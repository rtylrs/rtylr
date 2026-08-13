package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/roles/internal/domain"
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

func TestCreateRole_EmptyPermissions_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, "Empty Perms Role", []string{})
	if err != nil {
		t.Fatalf("roles.Create empty permissions: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestUpdateRole_PermissionsChange(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, "PermsChange Role", []string{"read"})
	if err != nil {
		t.Fatalf("roles.Create: %v", err)
	}

	if err := repo.Update(ctx, testutil.OwnerID, id, "PermsChange Role", []string{"read", "write", "delete"}); err != nil {
		t.Fatalf("roles.Update: %v", err)
	}

	roles, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("roles.List: %v", err)
	}
	for _, r := range roles {
		if r.ID == id && len(r.Permissions) != 3 {
			t.Errorf("expected 3 permissions after update, got %d", len(r.Permissions))
		}
	}
}

func TestDeleteRole_WrongOwner_NotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Delete with wrong owner and non-existent ID must report that nothing matched.
	err := repo.Delete(ctx, 9999999, 9999999)
	if !errors.Is(err, domain.ErrRoleNotFound) {
		t.Fatalf("roles.Delete(wrong owner, non-existent): expected ErrRoleNotFound, got %v", err)
	}
}

func TestCreateRole_Multiple_AllInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	names := []string{"Multi Role A", "Multi Role B", "Multi Role C"}
	ids := make([]int64, 0, len(names))
	for _, name := range names {
		id, err := repo.Create(ctx, testutil.OwnerID, name, []string{"read"})
		if err != nil {
			t.Fatalf("roles.Create %s: %v", name, err)
		}
		ids = append(ids, id)
	}

	roles, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("roles.List: %v", err)
	}
	found := make(map[int64]bool)
	for _, r := range roles {
		found[r.ID] = true
	}
	for _, id := range ids {
		if !found[id] {
			t.Errorf("role id=%d not found in list", id)
		}
	}
}

func TestUpdateRole_NotFound_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Update on a non-existent role must report that nothing matched.
	err := repo.Update(ctx, testutil.OwnerID, 9999999, "Ghost Role", []string{"read"})
	if !errors.Is(err, domain.ErrRoleNotFound) {
		t.Fatalf("roles.Update non-existent: expected ErrRoleNotFound, got %v", err)
	}
}

func TestCreateRole_UniqueIDsForDifferentNames(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.Create(ctx, testutil.OwnerID, "Unique Role X", []string{})
	if err != nil {
		t.Fatalf("roles.Create 1: %v", err)
	}
	id2, err := repo.Create(ctx, testutil.OwnerID, "Unique Role Y", []string{})
	if err != nil {
		t.Fatalf("roles.Create 2: %v", err)
	}
	if id1 == id2 {
		t.Errorf("expected distinct IDs for different roles, got %d and %d", id1, id2)
	}
}

func TestListRoles_ReturnedSliceNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	roles, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("roles.List: %v", err)
	}
	if roles == nil {
		t.Fatal("expected non-nil slice from List")
	}
}

func TestUpdateRole_NameAndPermissionsUpdated(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, "Original Role", []string{"view"})
	if err != nil {
		t.Fatalf("roles.Create: %v", err)
	}

	if err := repo.Update(ctx, testutil.OwnerID, id, "Renamed Role", []string{"view", "edit"}); err != nil {
		t.Fatalf("roles.Update: %v", err)
	}

	roles, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("roles.List: %v", err)
	}
	for _, r := range roles {
		if r.ID == id {
			if r.Name != "Renamed Role" {
				t.Errorf("expected name 'Renamed Role', got %s", r.Name)
			}
			if len(r.Permissions) != 2 {
				t.Errorf("expected 2 permissions, got %d", len(r.Permissions))
			}
			return
		}
	}
	t.Fatalf("updated role id=%d not found in list", id)
}

func TestListRoles_OwnerIDMatchesQuery(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, "OwnerID Check Role", []string{})
	if err != nil {
		t.Fatalf("roles.Create: %v", err)
	}
	_ = id

	roles, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("roles.List: %v", err)
	}
	for _, r := range roles {
		if r.OwnerID != testutil.OwnerID {
			t.Errorf("role id=%d has OwnerID=%d, want %d", r.ID, r.OwnerID, testutil.OwnerID)
		}
	}
}

func TestDeleteRole_ThenNotInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, "Lifecycle Role", []string{"read"})
	if err != nil {
		t.Fatalf("roles.Create: %v", err)
	}

	if err := repo.Delete(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("roles.Delete: %v", err)
	}

	roles, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("roles.List: %v", err)
	}
	for _, r := range roles {
		if r.ID == id {
			t.Errorf("deleted role id=%d still present in list", id)
		}
	}
}

func TestCreateRole_SinglePermission_Stored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, "Single Perm Role", []string{"admin"})
	if err != nil {
		t.Fatalf("roles.Create: %v", err)
	}

	roles, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("roles.List: %v", err)
	}
	for _, r := range roles {
		if r.ID == id {
			if len(r.Permissions) != 1 || r.Permissions[0] != "admin" {
				t.Errorf("expected permissions [admin], got %v", r.Permissions)
			}
			return
		}
	}
	t.Fatalf("role id=%d not found in list", id)
}
