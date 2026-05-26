package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/addons/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListAll_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	addons, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if addons == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestListAll_ItemsHaveNonEmptyID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	addons, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	for i, a := range addons {
		if a.ID == "" {
			t.Errorf("addons[%d].ID is empty", i)
		}
	}
}

func TestListAll_ItemsHaveNonEmptyName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	addons, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	for i, a := range addons {
		if a.Name == "" {
			t.Errorf("addons[%d].Name is empty", i)
		}
	}
}

func TestListForBusiness_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// business id=1 per fixtures
	addons, err := repo.ListForBusiness(ctx, 1)
	if err != nil {
		t.Fatalf("ListForBusiness: %v", err)
	}
	if addons == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestListForBusiness_UnknownBusinessReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	addons, err := repo.ListForBusiness(ctx, 999999)
	if err != nil {
		t.Fatalf("ListForBusiness unknown: %v", err)
	}
	if len(addons) != 0 {
		t.Fatalf("expected 0 addons for unknown business, got %d", len(addons))
	}
}

func TestListForBusiness_ItemsHaveAddonID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	addons, err := repo.ListForBusiness(ctx, 1)
	if err != nil {
		t.Fatalf("ListForBusiness: %v", err)
	}
	for i, a := range addons {
		if a.AddonID == "" {
			t.Errorf("addons[%d].AddonID is empty", i)
		}
	}
}

func TestEnable_KnownAddon_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Get any available addon ID
	all, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(all) == 0 {
		t.Skip("no addons in catalog, skipping enable test")
	}

	addonID := all[0].ID
	if err := repo.Enable(ctx, 1, addonID); err != nil {
		t.Fatalf("Enable: %v", err)
	}
}

func TestDisable_AfterEnable_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	all, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(all) == 0 {
		t.Skip("no addons in catalog, skipping disable test")
	}

	addonID := all[0].ID
	// Enable then disable
	_ = repo.Enable(ctx, 1, addonID)
	if err := repo.Disable(ctx, 1, addonID); err != nil {
		t.Fatalf("Disable: %v", err)
	}
}

func TestListForBusiness_ItemsHaveName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	addons, err := repo.ListForBusiness(ctx, 1)
	if err != nil {
		t.Fatalf("ListForBusiness: %v", err)
	}
	for i, a := range addons {
		if a.Name == "" {
			t.Errorf("addons[%d].Name is empty", i)
		}
	}
}

func TestListAll_PriceIsNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	addons, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	for i, a := range addons {
		if a.Price < 0 {
			t.Errorf("addons[%d].Price = %f, expected >= 0", i, a.Price)
		}
	}
}

func TestListAll_ConsistentTwoCalls(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll first: %v", err)
	}
	second, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll second: %v", err)
	}
	if len(first) != len(second) {
		t.Errorf("inconsistent results: first=%d second=%d", len(first), len(second))
	}
}
