package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/pos/tables/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/tables/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestList_ReturnsFixtureTables(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	tables, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tables) == 0 {
		t.Fatal("expected fixture rows, got none")
	}

	var foundT1, foundT2 bool
	for _, tbl := range tables {
		if tbl.ID == 3101 && tbl.Label == "T1" {
			foundT1 = true
		}
		if tbl.ID == 3102 && tbl.Label == "T2" {
			foundT2 = true
		}
	}
	if !foundT1 {
		t.Errorf("expected table id=3101 label='T1', not found in %+v", tables)
	}
	if !foundT2 {
		t.Errorf("expected table id=3102 label='T2', not found in %+v", tables)
	}
}

func TestList_Table3101IsAvailable(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	tables, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, tbl := range tables {
		if tbl.ID == 3101 {
			if tbl.Status != "available" {
				t.Errorf("expected table 3101 status='available', got %q", tbl.Status)
			}
			return
		}
	}
	t.Fatal("table id=3101 not found in fixture rows")
}

func TestListTables_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	tables, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(tables) != 0 {
		t.Errorf("expected empty slice for owner_id=0, got %d rows", len(tables))
	}
}

func TestListTables_ContainsT1(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	tables, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, tbl := range tables {
		if tbl.ID == 3101 && tbl.Label == "T1" {
			return
		}
	}
	t.Error("expected table id=3101 label='T1' in results")
}

func TestListTables_ContainsT2(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	tables, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, tbl := range tables {
		if tbl.ID == 3102 && tbl.Label == "T2" {
			return
		}
	}
	t.Error("expected table id=3102 label='T2' in results")
}

func TestListTables_T1Capacity4(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	tables, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, tbl := range tables {
		if tbl.ID == 3101 {
			if tbl.Seats != 4 {
				t.Errorf("expected T1 capacity=4, got %d", tbl.Seats)
			}
			return
		}
	}
	t.Fatal("table id=3101 not found")
}

func TestListTables_T2StatusOccupied(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	tables, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, tbl := range tables {
		if tbl.ID == 3102 {
			if tbl.Status != "occupied" {
				t.Errorf("expected T2 status='occupied', got %q", tbl.Status)
			}
			return
		}
	}
	t.Fatal("table id=3102 not found")
}

func TestCreateTable_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	locationID := int64(3001)
	input := domain.TableInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		LocationID: &locationID,
		Label:      "T99",
		Seats:      6,
	}

	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestCreateTable_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	locationID := int64(3001)
	input := domain.TableInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		LocationID: &locationID,
		Label:      "StoredTable",
		Seats:      2,
	}

	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	tables, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, tbl := range tables {
		if tbl.ID == id {
			if tbl.Label != "StoredTable" {
				t.Errorf("expected label='StoredTable', got %q", tbl.Label)
			}
			return
		}
	}
	t.Fatalf("created table id=%d not found in list", id)
}

func TestCreateTable_CapacityIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	locationID := int64(3001)
	input := domain.TableInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		LocationID: &locationID,
		Label:      "CapacityTable",
		Seats:      8,
	}

	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	tables, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, tbl := range tables {
		if tbl.ID == id {
			if tbl.Seats != 8 {
				t.Errorf("expected seats=8, got %d", tbl.Seats)
			}
			return
		}
	}
	t.Fatalf("created table id=%d not found in list", id)
}

func TestUpdateTable_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	locationID := int64(3001)
	id, err := repo.Create(ctx, domain.TableInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		LocationID: &locationID,
		Label:      "OldLabel",
		Seats:      4,
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	if err := repo.Update(ctx, domain.TableUpdateInput{
		ID:         id,
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		LocationID: &locationID,
		Label:      "NewLabel",
		Seats:      4,
		Status:     "available",
	}); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	tables, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, tbl := range tables {
		if tbl.ID == id {
			if tbl.Label != "NewLabel" {
				t.Errorf("expected label='NewLabel', got %q", tbl.Label)
			}
			return
		}
	}
	t.Fatalf("updated table id=%d not found in list", id)
}

func TestDeleteTable_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	locationID := int64(3001)
	id, err := repo.Create(ctx, domain.TableInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		LocationID: &locationID,
		Label:      "DeleteMe",
		Seats:      2,
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	if err := repo.Delete(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	tables, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, tbl := range tables {
		if tbl.ID == id {
			t.Errorf("expected table id=%d to be deleted, but still found in list", id)
		}
	}
}

func TestDeleteTable_NonExistentIsNoop(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.Delete(ctx, testutil.OwnerID, 999999999); err != nil {
		t.Errorf("expected no error deleting non-existent table, got: %v", err)
	}
}

func TestListTables_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	locationID := int64(3001)
	_, err = repo.Create(ctx, domain.TableInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		LocationID: &locationID,
		Label:      "ExtraTable",
		Seats:      3,
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	after, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	if len(after) <= len(before) {
		t.Errorf("expected count to increase after create; before=%d after=%d", len(before), len(after))
	}
}
