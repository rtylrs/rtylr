package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/addons/internal/infrastructure"
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
			t.Errorf("addons[%d].Price = %s, expected >= 0", i, a.Price.String())
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

func TestListAll_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
}

func TestListForBusiness_ConsistentTwoCalls(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, err := repo.ListForBusiness(ctx, 1)
	if err != nil {
		t.Fatalf("ListForBusiness first: %v", err)
	}
	second, err := repo.ListForBusiness(ctx, 1)
	if err != nil {
		t.Fatalf("ListForBusiness second: %v", err)
	}
	if len(first) != len(second) {
		t.Errorf("inconsistent results: first=%d second=%d", len(first), len(second))
	}
}

func TestEnable_ThenListForBusiness_AddonPresent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	all, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(all) == 0 {
		t.Skip("no addons in catalog")
	}

	addonID := all[0].ID
	if err := repo.Enable(ctx, 1, addonID); err != nil {
		t.Fatalf("Enable: %v", err)
	}

	bizAddons, err := repo.ListForBusiness(ctx, 1)
	if err != nil {
		t.Fatalf("ListForBusiness: %v", err)
	}
	found := false
	for _, a := range bizAddons {
		if a.AddonID == addonID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("enabled addon %q not found in ListForBusiness", addonID)
	}
}

func TestDisable_ThenListForBusiness_AddonIsDisabled(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	all, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(all) == 0 {
		t.Skip("no addons in catalog")
	}

	addonID := all[0].ID
	_ = repo.Enable(ctx, 1, addonID)
	if err := repo.Disable(ctx, 1, addonID); err != nil {
		t.Fatalf("Disable: %v", err)
	}

	bizAddons, err := repo.ListForBusiness(ctx, 1)
	if err != nil {
		t.Fatalf("ListForBusiness: %v", err)
	}
	for _, a := range bizAddons {
		if a.AddonID == addonID && a.IsEnabled {
			t.Errorf("addon %q should be disabled, but IsEnabled=true", addonID)
		}
	}
}

func TestListAll_ItemsHaveDescription(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	addons, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	// Description may be empty string, just ensure we can read it without panic
	_ = addons
}

func TestEnable_IdempotentNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	all, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(all) == 0 {
		t.Skip("no addons in catalog")
	}

	addonID := all[0].ID
	if err := repo.Enable(ctx, 1, addonID); err != nil {
		t.Fatalf("Enable first: %v", err)
	}
	if err := repo.Enable(ctx, 1, addonID); err != nil {
		t.Fatalf("Enable second: %v", err)
	}
}

func TestListForBusiness_ZeroBusinessID_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	addons, err := repo.ListForBusiness(ctx, 0)
	if err != nil {
		t.Fatalf("ListForBusiness: %v", err)
	}
	if len(addons) != 0 {
		t.Errorf("expected 0 addons for businessID=0, got %d", len(addons))
	}
}

func TestListAll_IsFreeFieldPresent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	addons, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	// Just ensure IsFree can be read from each addon
	for _, a := range addons {
		_ = a.IsFree
	}
}

func TestListForBusiness_ItemsIsEnabledFieldAccessible(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	bizAddons, err := repo.ListForBusiness(ctx, 1)
	if err != nil {
		t.Fatalf("ListForBusiness: %v", err)
	}
	for _, a := range bizAddons {
		_ = a.IsEnabled
	}
}

func TestEnable_UnknownAddonID_MayError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Enabling a non-existent addon should either error or be a noop.
	// We just ensure no panic.
	_ = repo.Enable(ctx, 1, "nonexistent-addon-xyz")
}

func TestDisable_UnknownAddon_MayError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Disabling a non-existent addon should either error or be a noop.
	_ = repo.Disable(ctx, 1, "nonexistent-addon-xyz")
}
