package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/price_lists/internal/domain"
	"github.com/voxire/onion/internal/modules/price_lists/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListPriceLists_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lists, err := repo.ListPriceLists(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListPriceLists: unexpected error: %v", err)
	}
	if lists == nil {
		t.Fatal("ListPriceLists: expected non-nil slice, got nil")
	}
}

func TestListPriceLists_WrongOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lists, err := repo.ListPriceLists(ctx, 0)
	if err != nil {
		t.Fatalf("ListPriceLists(ownerID=0): unexpected error: %v", err)
	}
	if len(lists) != 0 {
		t.Fatalf("ListPriceLists(ownerID=0): expected 0 rows, got %d", len(lists))
	}
}

func TestCreatePriceList_ReturnsInsertedID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.PriceListInput{
		Name:      "Test Price List",
		Currency:  "USD",
		IsDefault: false,
	}

	id, err := repo.CreatePriceList(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("CreatePriceList: unexpected error: %v", err)
	}
	if id <= 0 {
		t.Fatalf("CreatePriceList: expected positive ID, got %d", id)
	}
}

func TestCreatePriceList_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.PriceListInput{
		Name:      "Appears In List PL",
		Currency:  "EUR",
		IsDefault: false,
	}

	id, err := repo.CreatePriceList(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("CreatePriceList: %v", err)
	}

	lists, err := repo.ListPriceLists(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListPriceLists after create: %v", err)
	}

	var found bool
	for _, pl := range lists {
		if pl.ID == id {
			found = true
			if pl.Name != input.Name {
				t.Errorf("Name mismatch: got %q, want %q", pl.Name, input.Name)
			}
			if pl.Currency != input.Currency {
				t.Errorf("Currency mismatch: got %q, want %q", pl.Currency, input.Currency)
			}
			break
		}
	}
	if !found {
		t.Fatalf("created price list id=%d not found in list", id)
	}
}

func TestCreatePriceList_OwnerIDIsSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.PriceListInput{
		Name:     "Owner Check PL",
		Currency: "SAR",
	}

	id, err := repo.CreatePriceList(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("CreatePriceList: %v", err)
	}

	lists, err := repo.ListPriceLists(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListPriceLists: %v", err)
	}

	for _, pl := range lists {
		if pl.ID == id {
			if pl.OwnerID != testutil.OwnerID {
				t.Errorf("OwnerID mismatch: got %d, want %d", pl.OwnerID, testutil.OwnerID)
			}
			return
		}
	}
	t.Fatalf("created price list id=%d not found in list", id)
}

func TestCreatePriceList_DefaultFlag(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.PriceListInput{
		Name:      "Default Flag PL",
		Currency:  "USD",
		IsDefault: true,
	}

	id, err := repo.CreatePriceList(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("CreatePriceList: %v", err)
	}

	lists, err := repo.ListPriceLists(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListPriceLists: %v", err)
	}

	for _, pl := range lists {
		if pl.ID == id {
			if !pl.IsDefault {
				t.Errorf("expected IsDefault=true, got false")
			}
			return
		}
	}
	t.Fatalf("created price list id=%d not found in list", id)
}

func TestListPriceLists_EachHasNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Seed at least one
	_, err := repo.CreatePriceList(ctx, testutil.OwnerID, domain.PriceListInput{
		Name:     "ID Check PL",
		Currency: "USD",
	})
	if err != nil {
		t.Fatalf("CreatePriceList seed: %v", err)
	}

	lists, err := repo.ListPriceLists(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListPriceLists: %v", err)
	}

	for _, pl := range lists {
		if pl.ID <= 0 {
			t.Errorf("price list has non-positive ID: %d", pl.ID)
		}
	}
}

func TestListPriceLists_EachHasNonEmptyName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.CreatePriceList(ctx, testutil.OwnerID, domain.PriceListInput{
		Name:     "Name Check PL",
		Currency: "USD",
	})
	if err != nil {
		t.Fatalf("CreatePriceList seed: %v", err)
	}

	lists, err := repo.ListPriceLists(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListPriceLists: %v", err)
	}

	for _, pl := range lists {
		if pl.Name == "" {
			t.Errorf("price list id=%d has empty name", pl.ID)
		}
	}
}

func TestListPriceLists_EachHasCurrency(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.CreatePriceList(ctx, testutil.OwnerID, domain.PriceListInput{
		Name:     "Currency Check PL",
		Currency: "GBP",
	})
	if err != nil {
		t.Fatalf("CreatePriceList seed: %v", err)
	}

	lists, err := repo.ListPriceLists(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListPriceLists: %v", err)
	}

	for _, pl := range lists {
		if pl.Currency == "" {
			t.Errorf("price list id=%d has empty currency", pl.ID)
		}
	}
}

func TestCreateMultiplePriceLists_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.ListPriceLists(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListPriceLists before: %v", err)
	}

	for i := 0; i < 3; i++ {
		_, err := repo.CreatePriceList(ctx, testutil.OwnerID, domain.PriceListInput{
			Name:     "Bulk PL",
			Currency: "USD",
		})
		if err != nil {
			t.Fatalf("CreatePriceList[%d]: %v", i, err)
		}
	}

	after, err := repo.ListPriceLists(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListPriceLists after: %v", err)
	}

	if len(after) <= len(before) {
		t.Fatalf("expected list to grow: before=%d after=%d", len(before), len(after))
	}
}

func TestListPriceLists_ScopedToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Create for ownerID 1001
	_, err := repo.CreatePriceList(ctx, testutil.OwnerID, domain.PriceListInput{
		Name:     "Scoped PL",
		Currency: "USD",
	})
	if err != nil {
		t.Fatalf("CreatePriceList: %v", err)
	}

	// Query for a different owner should not see it
	other, err := repo.ListPriceLists(ctx, testutil.OwnerID+9999)
	if err != nil {
		t.Fatalf("ListPriceLists(other owner): %v", err)
	}
	for _, pl := range other {
		if pl.OwnerID == testutil.OwnerID {
			t.Errorf("result from other owner contains ownerID=%d item", testutil.OwnerID)
		}
	}
}
