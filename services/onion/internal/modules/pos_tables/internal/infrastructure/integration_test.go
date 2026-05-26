package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_tables/internal/infrastructure"
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
