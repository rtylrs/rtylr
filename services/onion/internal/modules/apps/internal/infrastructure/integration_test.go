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
	addons, _ := repo.GetEnabledAddons(ctx, 1)
	if addons == nil {
		t.Fatal("expected non-nil map")
	}
}

func TestGetEnabledAddons_UnknownBusinessReturnsEmptyMap(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	addons, _ := repo.GetEnabledAddons(ctx, 999999)
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

	addons, _ := repo.GetEnabledAddons(ctx, 1)
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

	addons, _ := repo.GetEnabledAddons(ctx, 1)
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
	apps, _ := repo.GetUserAllowedApps(ctx, 1, testutil.OwnerID)
	if apps == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestGetUserAllowedApps_UnknownUserReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	apps, _ := repo.GetUserAllowedApps(ctx, 1, 999999)
	if len(apps) != 0 {
		t.Fatalf("expected 0 apps for unknown user, got %d", len(apps))
	}
}

func TestGetUserAllowedApps_UnknownBusinessReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	apps, _ := repo.GetUserAllowedApps(ctx, 999999, testutil.OwnerID)
	// May return empty if no row found
	if apps == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestGetUserAllowedApps_AppStringsNonEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	apps, _ := repo.GetUserAllowedApps(ctx, 1, testutil.OwnerID)
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

	first, _ := repo.GetEnabledAddons(ctx, 1)
	second, _ := repo.GetEnabledAddons(ctx, 1)
	if len(first) != len(second) {
		t.Errorf("inconsistent results: first=%d second=%d", len(first), len(second))
	}
}

func TestGetUserAllowedApps_StaffUser(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// staff user 1002, business 1
	apps, _ := repo.GetUserAllowedApps(ctx, 1, 1002)
	if apps == nil {
		t.Fatal("expected non-nil slice for staff user")
	}
}

func TestGetUserAllowedApps_FallsBackForLegacyNullBusinessID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ownerID := int64(880001)
	res, err := db.ExecContext(ctx,
		"INSERT INTO businesses (name, owner_id) VALUES (?, ?)",
		"Legacy App Permissions", ownerID,
	)
	if err != nil {
		t.Fatalf("insert business: %v", err)
	}
	businessID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("business id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM businesses WHERE id = ?", businessID)
	})

	res, err = db.ExecContext(ctx,
		"INSERT INTO users (email, password_hash, full_name) VALUES (?, ?, ?)",
		"legacy-apps-member@example.com", "", "Legacy Apps Member",
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	userID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("user id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM users WHERE id = ?", userID)
	})

	_, err = db.ExecContext(ctx,
		"INSERT INTO user_app_permissions (business_id, owner_id, user_id, apps, role) VALUES (NULL, ?, ?, ?, ?)",
		ownerID, userID, `["pos","finance"]`, "staff",
	)
	if err != nil {
		t.Fatalf("insert legacy permission: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM user_app_permissions WHERE owner_id = ? AND user_id = ?", ownerID, userID)
	})

	apps, _ := repo.GetUserAllowedApps(ctx, businessID, userID)
	if len(apps) != 2 || apps[0] != "pos" || apps[1] != "finance" {
		t.Fatalf("expected legacy apps [pos finance], got %#v", apps)
	}
}

func TestGetEnabledAddons_ZeroBusinessID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	addons, _ := repo.GetEnabledAddons(ctx, 0)
	if addons == nil {
		t.Fatal("expected non-nil map even for zero business id")
	}
}

func TestGetUserAllowedApps_ChatUser(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// chat user 1003
	apps, _ := repo.GetUserAllowedApps(ctx, 1, 1003)
	if apps == nil {
		t.Fatal("expected non-nil slice for chat user")
	}
}

func TestGetEnabledAddons_ReturnsNonNilForBusiness1(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	result, _ := repo.GetEnabledAddons(ctx, 1)
	if result == nil {
		t.Fatal("expected non-nil map")
	}
}

func TestGetEnabledAddons_MultipleCalls_SameResult(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	a, _ := repo.GetEnabledAddons(ctx, 1)
	b, _ := repo.GetEnabledAddons(ctx, 1)
	if len(a) != len(b) {
		t.Errorf("results differ between calls: %d vs %d", len(a), len(b))
	}
}

func TestGetUserAllowedApps_NegativeUserID_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	apps, _ := repo.GetUserAllowedApps(ctx, 1, -1)
	if len(apps) != 0 {
		t.Errorf("expected 0 apps for negative user ID, got %d", len(apps))
	}
}

func TestGetEnabledAddons_NegativeBusinessID_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	result, _ := repo.GetEnabledAddons(ctx, -1)
	if result == nil {
		t.Fatal("expected non-nil map")
	}
	if len(result) != 0 {
		t.Errorf("expected 0 entries for negative businessID, got %d", len(result))
	}
}

func TestGetUserAllowedApps_SameBusiness_DifferentUsers(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	apps1, _ := repo.GetUserAllowedApps(ctx, 1, testutil.OwnerID)
	apps2, _ := repo.GetUserAllowedApps(ctx, 1, 1002)
	// Just ensure both return non-nil and don't panic
	if apps1 == nil || apps2 == nil {
		t.Fatal("expected non-nil slices for both users")
	}
}

func TestGetEnabledAddons_AllValuesTrue(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	addons, _ := repo.GetEnabledAddons(ctx, 1)
	for k, v := range addons {
		if !v {
			t.Errorf("addon %q has value false; only enabled addons should be in map", k)
		}
	}
}

func TestGetUserAllowedApps_ConsistentTwoCalls(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, _ := repo.GetUserAllowedApps(ctx, 1, testutil.OwnerID)
	second, _ := repo.GetUserAllowedApps(ctx, 1, testutil.OwnerID)
	if len(first) != len(second) {
		t.Errorf("inconsistent results: %d vs %d", len(first), len(second))
	}
}

func TestGetEnabledAddons_KeysAreStrings(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	addons, _ := repo.GetEnabledAddons(ctx, 1)
	for k := range addons {
		if k == "" {
			t.Error("found empty key in enabled addons map")
		}
	}
}

func TestGetUserAllowedApps_ZeroUserID_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	apps, _ := repo.GetUserAllowedApps(ctx, 1, 0)
	if len(apps) != 0 {
		t.Errorf("expected 0 apps for userID=0, got %d", len(apps))
	}
}

func TestGetEnabledAddons_UnknownBusinessReturnsEmptyNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	result, _ := repo.GetEnabledAddons(ctx, 888888)
	if result == nil {
		t.Fatal("expected non-nil map even for unknown business")
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}

func TestGetUserAllowedApps_AppStringsAreNonEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	apps, _ := repo.GetUserAllowedApps(ctx, 1, testutil.OwnerID)
	for i, a := range apps {
		if a == "" {
			t.Errorf("apps[%d] is empty string", i)
		}
	}
}

func TestGetEnabledAddons_BusinessZero_ReturnsNonNilMap(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	result, _ := repo.GetEnabledAddons(ctx, 0)
	if result == nil {
		t.Fatal("expected non-nil map for businessID=0")
	}
}
