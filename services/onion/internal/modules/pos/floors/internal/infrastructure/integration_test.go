package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/pos/floors/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/floors/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const (
	floorsBusinessID int64 = 1
	floorsBranchID   int64 = 3001
)

func TestListFloors_RunsWithoutError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.List(ctx, floorsBusinessID, floorsBranchID)
	if err != nil {
		t.Fatalf("unexpected error listing floors for business %d branch %d: %v", floorsBusinessID, floorsBranchID, err)
	}
}

func TestListFloors_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	floors, err := repo.List(ctx, 0, 0)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(floors) != 0 {
		t.Errorf("expected empty slice for businessID=0, got %d rows", len(floors))
	}
}

func TestListFloors_IsNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	floors, err := repo.List(ctx, floorsBusinessID, floorsBranchID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if floors == nil {
		t.Error("expected non-nil slice, got nil")
	}
}

func TestListFloors_NotNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	result, err := repo.List(ctx, floorsBusinessID, floorsBranchID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	// Re-assert non-nil to act as a standalone guard.
	if result == nil {
		t.Error("List returned nil slice")
	}
}

func TestCreateFloor_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.FloorInput{
		BusinessID: floorsBusinessID,
		BranchID:   floorsBranchID,
		Name:       "Ground Floor",
		SortOrder:  1,
	}

	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestCreateFloor_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.FloorInput{
		BusinessID: floorsBusinessID,
		BranchID:   floorsBranchID,
		Name:       "Mezzanine",
		SortOrder:  2,
	}

	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	floors, err := repo.List(ctx, floorsBusinessID, floorsBranchID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, f := range floors {
		if f.ID == id {
			if f.Name != "Mezzanine" {
				t.Errorf("expected name='Mezzanine', got %q", f.Name)
			}
			return
		}
	}
	t.Fatalf("created floor id=%d not found in list", id)
}

func TestCreateFloor_BranchIDIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.FloorInput{
		BusinessID: floorsBusinessID,
		BranchID:   floorsBranchID,
		Name:       "BranchFloor",
		SortOrder:  3,
	}

	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	floors, err := repo.List(ctx, floorsBusinessID, floorsBranchID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, f := range floors {
		if f.ID == id {
			if f.BranchID != floorsBranchID {
				t.Errorf("expected BranchID=%d, got %d", floorsBranchID, f.BranchID)
			}
			return
		}
	}
	t.Fatalf("created floor id=%d not found in list", id)
}

func TestUpdateFloor_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.FloorInput{
		BusinessID: floorsBusinessID,
		BranchID:   floorsBranchID,
		Name:       "OldName",
		SortOrder:  1,
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	if err := repo.Update(ctx, domain.FloorUpdateInput{
		ID:         id,
		BusinessID: floorsBusinessID,
		Name:       "NewName",
		SortOrder:  1,
	}); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	floors, err := repo.List(ctx, floorsBusinessID, floorsBranchID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, f := range floors {
		if f.ID == id {
			if f.Name != "NewName" {
				t.Errorf("expected name='NewName', got %q", f.Name)
			}
			return
		}
	}
	t.Fatalf("updated floor id=%d not found in list", id)
}

func TestDeleteFloor_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.FloorInput{
		BusinessID: floorsBusinessID,
		BranchID:   floorsBranchID,
		Name:       "ToDelete",
		SortOrder:  1,
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	if err := repo.Delete(ctx, floorsBusinessID, id); err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	floors, err := repo.List(ctx, floorsBusinessID, floorsBranchID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, f := range floors {
		if f.ID == id {
			t.Errorf("expected floor id=%d to be deleted, but still found in list", id)
		}
	}
}

func TestDeleteFloor_NonExistentIsNoop(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.Delete(ctx, floorsBusinessID, 999999999); err != nil {
		t.Errorf("expected no error deleting non-existent floor, got: %v", err)
	}
}

func TestListFloors_AfterCreate_ContainsNew(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.FloorInput{
		BusinessID: floorsBusinessID,
		BranchID:   floorsBranchID,
		Name:       "NewlyAdded",
		SortOrder:  10,
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	floors, err := repo.List(ctx, floorsBusinessID, floorsBranchID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, f := range floors {
		if f.ID == id {
			return
		}
	}
	t.Fatalf("newly created floor id=%d not found in list", id)
}

func TestCreate_ThenList_FindsFloor(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.FloorInput{
		BusinessID: floorsBusinessID,
		BranchID:   floorsBranchID,
		Name:       "RoundtripFloor",
		SortOrder:  5,
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	floors, err := repo.List(ctx, floorsBusinessID, floorsBranchID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	found := false
	for _, f := range floors {
		if f.ID == id && f.Name == "RoundtripFloor" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("roundtrip floor id=%d with name='RoundtripFloor' not found", id)
	}
}
