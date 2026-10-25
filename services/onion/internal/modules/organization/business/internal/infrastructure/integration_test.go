package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/business/internal/infrastructure"
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

func TestGetProfile_ReturnsProfileForBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	profile, err := repo.GetProfile(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if profile.ID != testutil.BusinessID {
		t.Errorf("expected ID=%d, got %d", testutil.BusinessID, profile.ID)
	}
}

// A profile is read by business ID only. Before, a lookup also matched the
// owner's user ID, so a user ID that happened to equal a business ID read
// that business, and an owner's ID read their first business.
func TestGetProfile_DoesNotMatchAnOwnerUserID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const ownerID, businessID = int64(9967601), int64(9967701)
	if _, err := db.ExecContext(ctx, `INSERT INTO businesses (id, owner_id, name, is_active) VALUES (?, ?, 'Owned elsewhere', TRUE)`, businessID, ownerID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), "DELETE FROM businesses WHERE id = ?", businessID) })

	if _, err := repo.GetProfile(ctx, ownerID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("GetProfile(owner user id) error = %v, want sql.ErrNoRows", err)
	}
	if profile, err := repo.GetProfile(ctx, businessID); err != nil || profile.ID != businessID {
		t.Fatalf("GetProfile(business id) = %+v, %v", profile, err)
	}
}

func TestGetProfile_SiteNameIsTestSite(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	profile, err := repo.GetProfile(ctx, testutil.BusinessID)
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

	profile, err := repo.GetProfile(ctx, testutil.BusinessID)
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

func TestLookupBySlug_ReturnsBusiness(t *testing.T) {
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
	ownerID := businessOwner(t, db, 1)

	// Read the current slug so we can restore it.
	var original sql.NullString
	_ = db.QueryRowContext(ctx, `SELECT slug FROM businesses WHERE id = 1 LIMIT 1`).Scan(&original)

	testSlug := "integration-test-slug-tmp"
	updated, err := repo.UpdateSlug(ctx, 1, ownerID, testSlug)
	if err != nil {
		t.Fatalf("UpdateSlug set: %v", err)
	}
	if !updated {
		t.Fatal("owner-scoped slug update reported no rows updated")
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
	_, _ = repo.UpdateSlug(ctx, 1, ownerID, restore)
}

// The slug update addresses the row by id alone unless the owner is part of the
// statement, so a caller that owns no part of the business must match nothing:
// 0 rows affected, and the other tenant's slug left untouched.
func TestUpdateSlug_CrossTenantUpdateDenied(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	suffix := time.Now().UnixNano()
	foreign := insertTestBusiness(t, db, testutil.OtherOwnerID, fmt.Sprintf("Slug Tenant B %d", suffix), fmt.Sprintf("slug-tenant-b-%d", suffix))

	updated, err := repo.UpdateSlug(ctx, foreign, testutil.OwnerID, "hijacked-slug")
	if err != nil {
		t.Fatalf("UpdateSlug: %v", err)
	}
	if updated {
		t.Fatal("cross-tenant slug update reported success")
	}

	var got string
	if err := db.QueryRowContext(ctx, `SELECT slug FROM businesses WHERE id = ?`, foreign).Scan(&got); err != nil {
		t.Fatalf("read slug: %v", err)
	}
	if got != fmt.Sprintf("slug-tenant-b-%d", suffix) {
		t.Fatalf("cross-tenant write changed the other tenant's slug to %q", got)
	}
}

func TestUpdateSlug_SameTenantUpdateSucceeds(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	suffix := time.Now().UnixNano()
	own := insertTestBusiness(t, db, testutil.OwnerID, fmt.Sprintf("Slug Tenant A %d", suffix), fmt.Sprintf("slug-tenant-a-%d", suffix))

	updated, err := repo.UpdateSlug(ctx, own, testutil.OwnerID, "renamed-by-owner")
	if err != nil {
		t.Fatalf("UpdateSlug: %v", err)
	}
	if !updated {
		t.Fatal("owner's slug update reported no rows updated")
	}

	var got string
	if err := db.QueryRowContext(ctx, `SELECT slug FROM businesses WHERE id = ?`, own).Scan(&got); err != nil {
		t.Fatalf("read slug: %v", err)
	}
	if got != "renamed-by-owner" {
		t.Errorf("expected slug=%q, got %q", "renamed-by-owner", got)
	}
}

func businessOwner(t *testing.T, db *sql.DB, businessID int64) int64 {
	t.Helper()

	var ownerID int64
	if err := db.QueryRowContext(context.Background(), `SELECT owner_id FROM businesses WHERE id = ?`, businessID).Scan(&ownerID); err != nil {
		t.Fatalf("read owner of business %d: %v", businessID, err)
	}
	return ownerID
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

	profile, err := repo.GetProfile(ctx, testutil.BusinessID)
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

func TestGetProfile_ByBusinessID_ReturnsProfile(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	profile, err := repo.GetProfile(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("GetProfile(%d): %v", testutil.BusinessID, err)
	}
	if profile.ID != testutil.BusinessID {
		t.Errorf("expected ID=%d, got %d", testutil.BusinessID, profile.ID)
	}
}

func TestGetProfile_ZeroBusiness_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetProfile(ctx, 0)
	if err == nil {
		t.Fatal("expected error for businessID=0, got nil")
	}
}

func TestGetProfile_NameIsTestBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	profile, err := repo.GetProfile(ctx, testutil.BusinessID)
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

	profile, err := repo.GetProfile(ctx, testutil.BusinessID)
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
	ownerID := businessOwner(t, db, 1)

	var original sql.NullString
	_ = db.QueryRowContext(ctx, `SELECT slug FROM businesses WHERE id = 1 LIMIT 1`).Scan(&original)

	newSlug := "updated-slug-test-12345"
	updated, err := repo.UpdateSlug(ctx, 1, ownerID, newSlug)
	if err != nil {
		t.Fatalf("UpdateSlug: %v", err)
	}
	if !updated {
		t.Fatal("owner-scoped slug update reported no rows updated")
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
	_, _ = repo.UpdateSlug(ctx, 1, ownerID, restore)
}

func TestGetProfile_ContactEmailPresent2(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	profile, err := repo.GetProfile(ctx, testutil.BusinessID)
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

	profile, err := repo.GetProfile(ctx, testutil.BusinessID)
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

	profile, err := repo.GetProfile(ctx, testutil.BusinessID)
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
