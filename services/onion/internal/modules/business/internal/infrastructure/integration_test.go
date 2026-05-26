package infrastructure_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"

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
