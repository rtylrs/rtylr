package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_locations/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_locations/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestLocations_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	locs, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	_ = locs
}

func TestLocations_List_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	locs, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if locs == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
}

func TestLocations_List_ContainsTestBranch(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	locs, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	found := false
	for _, l := range locs {
		if l.ID == 3001 && l.Name == "Test Branch" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected location id=3001 name='Test Branch' in result, got %+v", locs)
	}
}

func TestLocations_List_TestBranchIsActive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	locs, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	for _, l := range locs {
		if l.ID == 3001 {
			if !l.IsActive {
				t.Errorf("location 3001 expected IsActive=true, got false")
			}
			return
		}
	}
	t.Error("location 3001 not found")
}

func TestLocations_List_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	locs, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, l := range locs {
		if l.OwnerID != testutil.OwnerID {
			t.Errorf("locs[%d].OwnerID = %d, want %d", i, l.OwnerID, testutil.OwnerID)
		}
	}
}

func TestLocations_List_AllHaveNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	locs, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, l := range locs {
		if l.ID == 0 {
			t.Errorf("locs[%d].ID is zero", i)
		}
	}
}

func TestLocations_List_AllHaveNonEmptyName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	locs, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, l := range locs {
		if l.Name == "" {
			t.Errorf("locs[%d].Name is empty", i)
		}
	}
}

func TestLocations_List_WrongOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	locs, err := repo.List(context.Background(), 999999)
	if err != nil {
		t.Fatalf("List wrong owner: %v", err)
	}
	if len(locs) != 0 {
		t.Fatalf("expected 0 locations for unknown owner, got %d", len(locs))
	}
}

func TestLocations_List_AtLeastOneLocation(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	locs, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(locs) == 0 {
		t.Fatal("expected at least one location for fixture owner")
	}
}

func TestLocations_Create_ThenListContainsNew(t *testing.T) {
	db := testutil.OpenDB(t)
	testutil.GrantBusinessAccess(t, db, testutil.OtherBusinessID, testutil.OtherOwnerID)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	newID, err := repo.Create(ctx, domain.LocationInput{
		OwnerID:    testutil.OtherOwnerID,
		BusinessID: testutil.OtherBusinessID,
		Name:       "Integration Test Location",
		IsActive:   true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if newID == 0 {
		t.Fatal("Create returned ID=0")
	}

	locs, err := repo.List(ctx, testutil.OtherBusinessID)
	if err != nil {
		t.Fatalf("List after create: %v", err)
	}
	found := false
	for _, l := range locs {
		if l.ID == newID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("newly created location id=%d not found in list", newID)
	}
}

func TestLocations_Update_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	testutil.GrantBusinessAccess(t, db, testutil.OtherBusinessID, testutil.OtherOwnerID)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	active := true

	newID, err := repo.Create(ctx, domain.LocationInput{
		OwnerID:    testutil.OtherOwnerID,
		BusinessID: testutil.OtherBusinessID,
		Name:       "Before Update",
		IsActive:   true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = repo.Update(ctx, domain.LocationUpdateInput{
		ID:         newID,
		OwnerID:    testutil.OtherOwnerID,
		BusinessID: testutil.OtherBusinessID,
		Name:       "After Update",
		IsActive:   &active,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	locs, err := repo.List(ctx, testutil.OtherBusinessID)
	if err != nil {
		t.Fatalf("List after update: %v", err)
	}
	for _, l := range locs {
		if l.ID == newID {
			if l.Name != "After Update" {
				t.Errorf("expected name 'After Update', got %q", l.Name)
			}
			return
		}
	}
	t.Errorf("updated location id=%d not found", newID)
}

func TestLocations_Update_UnknownIDReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	active := true

	err := repo.Update(ctx, domain.LocationUpdateInput{
		ID:         9999999,
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Ghost",
		IsActive:   &active,
	})
	if err == nil {
		t.Fatal("expected error for unknown location, got nil")
	}
}

func TestLocations_List_TestBranchOwnerID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	locs, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, l := range locs {
		if l.ID == 3001 {
			if l.OwnerID != testutil.OwnerID {
				t.Errorf("location 3001 OwnerID = %d, want %d", l.OwnerID, testutil.OwnerID)
			}
			return
		}
	}
	t.Error("location 3001 not found")
}
