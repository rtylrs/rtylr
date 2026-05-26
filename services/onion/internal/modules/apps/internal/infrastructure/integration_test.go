package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/apps/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetEnabledAddons_ReturnsMap(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// business id=1 per fixtures
	addons := repo.GetEnabledAddons(ctx, 1)
	if addons == nil {
		t.Fatal("expected non-nil map")
	}
}

func TestGetEnabledAddons_UnknownBusinessReturnsEmptyMap(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	addons := repo.GetEnabledAddons(ctx, 999999)
	if addons == nil {
		t.Fatal("expected non-nil map, got nil")
	}
	if len(addons) != 0 {
		t.Fatalf("expected empty map for unknown business, got %d entries", len(addons))
	}
}

func TestGetEnabledAddons_ValuesAreTrue(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	addons := repo.GetEnabledAddons(ctx, 1)
	for k, v := range addons {
		if !v {
			t.Errorf("addon %q has value false, expected true", k)
		}
	}
}

func TestGetEnabledAddons_KeysNonEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	addons := repo.GetEnabledAddons(ctx, 1)
	for k := range addons {
		if k == "" {
			t.Error("addon map contains empty string key")
		}
	}
}

func TestGetUserAllowedApps_ReturnsSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// owner user 1001, business 1
	apps := repo.GetUserAllowedApps(ctx, 1, testutil.OwnerID)
	if apps == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestGetUserAllowedApps_UnknownUserReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	apps := repo.GetUserAllowedApps(ctx, 1, 999999)
	if len(apps) != 0 {
		t.Fatalf("expected 0 apps for unknown user, got %d", len(apps))
	}
}

func TestGetUserAllowedApps_UnknownBusinessReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	apps := repo.GetUserAllowedApps(ctx, 999999, testutil.OwnerID)
	// May return empty if no row found
	if apps == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestGetUserAllowedApps_AppStringsNonEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	apps := repo.GetUserAllowedApps(ctx, 1, testutil.OwnerID)
	for i, a := range apps {
		if a == "" {
			t.Errorf("apps[%d] is empty string", i)
		}
	}
}

func TestGetEnabledAddons_ConsistentTwoCalls(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first := repo.GetEnabledAddons(ctx, 1)
	second := repo.GetEnabledAddons(ctx, 1)
	if len(first) != len(second) {
		t.Errorf("inconsistent results: first=%d second=%d", len(first), len(second))
	}
}

func TestGetUserAllowedApps_StaffUser(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// staff user 1002, business 1
	apps := repo.GetUserAllowedApps(ctx, 1, 1002)
	if apps == nil {
		t.Fatal("expected non-nil slice for staff user")
	}
}

func TestGetEnabledAddons_ZeroBusinessID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	addons := repo.GetEnabledAddons(ctx, 0)
	if addons == nil {
		t.Fatal("expected non-nil map even for zero business id")
	}
}

func TestGetUserAllowedApps_ChatUser(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// chat user 1003
	apps := repo.GetUserAllowedApps(ctx, 1, 1003)
	if apps == nil {
		t.Fatal("expected non-nil slice for chat user")
	}
}
