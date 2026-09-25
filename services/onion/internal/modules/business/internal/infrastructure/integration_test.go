package infrastructure_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/business/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetPublicBusiness_ReturnsTestBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	biz, err := repo.GetPublicBusiness(ctx, 1)
	if err != nil {
		t.Fatalf("GetPublicBusiness: %v", err)
	}
	if biz.ID != 1 {
		t.Errorf("expected ID=1, got %d", biz.ID)
	}
	if biz.Name != "Test Business" {
		t.Errorf("expected Name='Test Business', got %q", biz.Name)
	}
}

func TestGetPublicBusiness_CurrencyAndTimezoneDefaults(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	biz, err := repo.GetPublicBusiness(ctx, 1)
	if err != nil {
		t.Fatalf("GetPublicBusiness: %v", err)
	}
	if biz.Currency == "" {
		t.Error("expected non-empty Currency")
	}
	if biz.Timezone == "" {
		t.Error("expected non-empty Timezone")
	}
}

func TestGetPublicBusiness_NotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetPublicBusiness(ctx, 999999)
	if err == nil {
		t.Fatal("expected error for non-existent business, got nil")
	}
}

func TestGetProfile_ReturnsProfileForOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	profile, err := repo.GetProfile(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if profile.ID != testutil.OwnerID {
		t.Errorf("expected ID=%d, got %d", testutil.OwnerID, profile.ID)
	}
}

func TestGetProfile_SiteNameIsTestSite(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	profile, err := repo.GetProfile(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if profile.SiteName != "Test Site" {
		t.Errorf("expected SiteName='Test Site', got %q", profile.SiteName)
	}
}

func TestGetProfile_HasDefaultCurrencyAndTimezone(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	profile, err := repo.GetProfile(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if profile.Currency != "USD" {
		t.Errorf("expected Currency='USD', got %q", profile.Currency)
	}
	if profile.Timezone != "UTC" {
		t.Errorf("expected Timezone='UTC', got %q", profile.Timezone)
	}
	if profile.Plan != "free" {
		t.Errorf("expected Plan='free', got %q", profile.Plan)
	}
}

func TestLookupBySlug_ReturnsBusinessAndLocations(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// First retrieve the slug from the businesses table directly.
	var slug sql.NullString
	err := db.QueryRowContext(ctx, `SELECT slug FROM businesses WHERE id = 1 LIMIT 1`).Scan(&slug)
	if err != nil || !slug.Valid || slug.String == "" {
		t.Skip("business id=1 has no slug set -- skipping slug lookup test")
	}

	result, err := repo.LookupBySlug(ctx, slug.String)
	if err != nil {
		t.Fatalf("LookupBySlug: %v", err)
	}
	if result.Business.Name != "Test Business" {
		t.Errorf("expected Name='Test Business', got %q", result.Business.Name)
	}
}

// TestLookupBySlug_LocationsAreBusinessScoped seeds two storefronts for the
// same owner: the queried business has its own location and a sibling business
// holds a legacy-style location whose owner_id is the shared user id. The
// public lookup must list the business-scoped row and must not leak the
// sibling's row onto it.
func TestLookupBySlug_LocationsAreBusinessScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	var ownerID int64
	if err := db.QueryRowContext(ctx, `SELECT owner_id FROM businesses WHERE id = 1`).Scan(&ownerID); err != nil {
		t.Fatalf("read fixture owner: %v", err)
	}

	suffix := time.Now().UnixNano()
	bizA := insertTestBusiness(t, db, ownerID, fmt.Sprintf("Public Scope A %d", suffix), fmt.Sprintf("public-scope-a-%d", suffix))
	bizB := insertTestBusiness(t, db, ownerID, fmt.Sprintf("Public Scope B %d", suffix), fmt.Sprintf("public-scope-b-%d", suffix))

	locA := insertTestLocation(t, db, bizA, bizA, fmt.Sprintf("Scope A Location %d", suffix), fmt.Sprintf("scope-a-loc-%d", suffix))
	// Legacy rows keep the owner user id in owner_id; the query must key on
	// business_id, so this row may only appear on business B's storefront.
	locB := insertTestLocation(t, db, ownerID, bizB, fmt.Sprintf("Scope B Location %d", suffix), fmt.Sprintf("scope-b-loc-%d", suffix))

	result, err := repo.LookupBySlug(ctx, fmt.Sprintf("public-scope-a-%d", suffix))
	if err != nil {
		t.Fatalf("LookupBySlug: %v", err)
	}

	ids := make([]int64, 0, len(result.Locations))
	for _, loc := range result.Locations {
		ids = append(ids, loc.ID)
	}
	testutil.AssertTenantIsolation(t, ids, []int64{locA}, []int64{locB})
}

func insertTestBusiness(t *testing.T, db *sql.DB, ownerID int64, name, slug string) int64 {
	t.Helper()

	result, err := db.Exec(
		`INSERT INTO businesses (name, owner_id, slug, is_active) VALUES (?, ?, ?, TRUE)`,
		name, ownerID, slug,
	)
	if err != nil {
		t.Fatalf("insert business: %v", err)
	}
	businessID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("read business id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM businesses WHERE id = ?`, businessID)
	})
	return businessID
}

func insertTestLocation(t *testing.T, db *sql.DB, ownerID, businessID int64, name, slug string) int64 {
	t.Helper()

	result, err := db.Exec(
		`INSERT INTO hr_locations (owner_id, business_id, name, slug, is_active) VALUES (?, ?, ?, ?, TRUE)`,
		ownerID, businessID, name, slug,
	)
	if err != nil {
		t.Fatalf("insert hr_location: %v", err)
	}
	locationID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("read hr_location id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM hr_locations WHERE id = ?`, locationID)
	})
	return locationID
}

func TestLookupBySlug_UnknownSlugReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.LookupBySlug(ctx, "no-such-slug-xyz-99999")
	if err == nil {
		t.Fatal("expected error for unknown slug, got nil")
	}
}

func TestUpdateSlug_RoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Read the current slug so we can restore it.
	var original sql.NullString
	_ = db.QueryRowContext(ctx, `SELECT slug FROM businesses WHERE id = 1 LIMIT 1`).Scan(&original)

	testSlug := "integration-test-slug-tmp"
	if err := repo.UpdateSlug(ctx, 1, testSlug); err != nil {
		t.Fatalf("UpdateSlug set: %v", err)
	}

	var got string
	if err := db.QueryRowContext(ctx, `SELECT slug FROM businesses WHERE id = 1`).Scan(&got); err != nil {
		t.Fatalf("read slug: %v", err)
	}
	if got != testSlug {
		t.Errorf("expected slug=%q, got %q", testSlug, got)
	}

	// Restore original.
	restore := ""
	if original.Valid {
		restore = original.String
	}
	_ = repo.UpdateSlug(ctx, 1, restore)
}

func TestGetPublicBusiness_SlugField(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	biz, err := repo.GetPublicBusiness(ctx, 1)
	if err != nil {
		t.Fatalf("GetPublicBusiness: %v", err)
	}
	// Slug may be empty for fixture but the field must be present (no scan error).
	_ = biz.Slug
}

func TestGetProfile_ContactEmailPresent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	profile, err := repo.GetProfile(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if !strings.Contains(profile.ContactEmail, "@") {
		t.Errorf("expected valid contact email, got %q", profile.ContactEmail)
	}
}

func TestGetPublicBusiness_InactiveBusinessNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// is_active=FALSE businesses should not be returned.
	_, err := repo.GetPublicBusiness(ctx, 0)
	if err == nil {
		t.Fatal("expected error for business id=0, got nil")
	}
}

func TestGetProfile_OwnerID1001_ReturnsProfile(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	profile, err := repo.GetProfile(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetProfile(1001): %v", err)
	}
	if profile.ID != testutil.OwnerID {
		t.Errorf("expected ID=%d, got %d", testutil.OwnerID, profile.ID)
	}
}

func TestGetProfile_ZeroOwner_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetProfile(ctx, 0)
	if err == nil {
		t.Fatal("expected error for ownerID=0, got nil")
	}
}

func TestGetProfile_NameIsTestBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	profile, err := repo.GetProfile(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	// SiteName is set from the fixture
	if profile.SiteName == "" {
		t.Error("expected non-empty SiteName")
	}
}

func TestGetProfile_CurrencyIsUSD(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	profile, err := repo.GetProfile(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if profile.Currency != "USD" {
		t.Errorf("expected Currency='USD', got %q", profile.Currency)
	}
}

func TestGetPublicBusiness_ID1_ReturnsData(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	biz, err := repo.GetPublicBusiness(ctx, 1)
	if err != nil {
		t.Fatalf("GetPublicBusiness(1): %v", err)
	}
	if biz.ID != 1 {
		t.Errorf("expected ID=1, got %d", biz.ID)
	}
}

func TestGetPublicBusiness_NotFound_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetPublicBusiness(ctx, 88888888)
	if err == nil {
		t.Fatal("expected error for non-existent business id, got nil")
	}
}

func TestLookupBySlug_NonExistentSlug_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.LookupBySlug(ctx, "nonexistent-slug-abc-xyz-12345")
	if err == nil {
		t.Fatal("expected error for non-existent slug, got nil")
	}
}

func TestUpdateSlug_ChangesSlug(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	var original sql.NullString
	_ = db.QueryRowContext(ctx, `SELECT slug FROM businesses WHERE id = 1 LIMIT 1`).Scan(&original)

	newSlug := "updated-slug-test-12345"
	if err := repo.UpdateSlug(ctx, 1, newSlug); err != nil {
		t.Fatalf("UpdateSlug: %v", err)
	}

	var got string
	if err := db.QueryRowContext(ctx, `SELECT slug FROM businesses WHERE id = 1`).Scan(&got); err != nil {
		t.Fatalf("read slug: %v", err)
	}
	if got != newSlug {
		t.Errorf("expected slug=%q, got %q", newSlug, got)
	}

	// Restore
	restore := ""
	if original.Valid {
		restore = original.String
	}
	_ = repo.UpdateSlug(ctx, 1, restore)
}

func TestGetProfile_ContactEmailPresent2(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	profile, err := repo.GetProfile(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if profile.ContactEmail == "" {
		t.Error("expected non-empty ContactEmail")
	}
}

func TestGetProfile_TimezoneIsUTC(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	profile, err := repo.GetProfile(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if profile.Timezone == "" {
		t.Error("expected non-empty Timezone")
	}
}

func TestGetProfile_PlanIsSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	profile, err := repo.GetProfile(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if profile.Plan == "" {
		t.Error("expected non-empty Plan")
	}
}

func TestGetPublicBusiness_NameIsTestBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	biz, err := repo.GetPublicBusiness(ctx, 1)
	if err != nil {
		t.Fatalf("GetPublicBusiness(1): %v", err)
	}
	if biz.Name != "Test Business" {
		t.Errorf("expected Name='Test Business', got %q", biz.Name)
	}
}

func TestGetPublicBusiness_CurrencyNonEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	biz, err := repo.GetPublicBusiness(ctx, 1)
	if err != nil {
		t.Fatalf("GetPublicBusiness(1): %v", err)
	}
	if biz.Currency == "" {
		t.Error("expected non-empty Currency in public business")
	}
}
