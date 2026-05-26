package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/config/internal/domain"
	"github.com/voxire/onion/internal/modules/config/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
	"strconv"
)

func TestGetConfig_ReturnsSiteNameTestSite(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ownerID := strconv.FormatInt(testutil.OwnerID, 10)
	cfg, err := repo.GetConfig(ctx, ownerID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.SiteName != "Test Site" {
		t.Errorf("expected SiteName='Test Site', got %q", cfg.SiteName)
	}
}

func TestGetConfig_ContactEmailPresent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cfg, err := repo.GetConfig(ctx, strconv.FormatInt(testutil.OwnerID, 10))
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.ContactEmail == "" {
		t.Error("expected non-empty ContactEmail")
	}
}

func TestGetConfig_OwnerIDMatchesRequest(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ownerStr := strconv.FormatInt(testutil.OwnerID, 10)
	cfg, err := repo.GetConfig(ctx, ownerStr)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.OwnerId != ownerStr {
		t.Errorf("expected OwnerId=%q, got %q", ownerStr, cfg.OwnerId)
	}
}

func TestGetConfig_InvalidOwnerIDReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetConfig(ctx, "not-a-number")
	if err == nil {
		t.Fatal("expected error for non-numeric ownerID, got nil")
	}
}

func TestGetConfig_UnknownOwnerReturnsErrNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetConfig(ctx, "999888777")
	if err == nil {
		t.Fatal("expected error for unknown owner, got nil")
	}
}

func TestGetThemeByOwnerID_NoErrorForOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// May return empty theme if none set; must not error.
	_, err := repo.GetThemeByOwnerID(ctx, strconv.FormatInt(testutil.OwnerID, 10))
	if err != nil {
		t.Fatalf("GetThemeByOwnerID: %v", err)
	}
}

func TestGetThemeByOwnerID_InvalidOwnerReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetThemeByOwnerID(ctx, "bad-id")
	if err == nil {
		t.Fatal("expected error for non-numeric ownerID, got nil")
	}
}

func TestUpdateConfig_RoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ownerStr := strconv.FormatInt(testutil.OwnerID, 10)

	original, err := repo.GetConfig(ctx, ownerStr)
	if err != nil {
		t.Fatalf("GetConfig (original): %v", err)
	}

	input := domain.ConfigInput{
		OwnerId:      ownerStr,
		SiteName:     "Updated Site",
		ContactEmail: original.ContactEmail,
	}
	if err := repo.UpdateConfig(ctx, input); err != nil {
		t.Fatalf("UpdateConfig: %v", err)
	}

	updated, err := repo.GetConfig(ctx, ownerStr)
	if err != nil {
		t.Fatalf("GetConfig (after update): %v", err)
	}
	if updated.SiteName != "Updated Site" {
		t.Errorf("expected SiteName='Updated Site', got %q", updated.SiteName)
	}

	// Restore original.
	restore := domain.ConfigInput{
		OwnerId:      ownerStr,
		SiteName:     original.SiteName,
		ContactEmail: original.ContactEmail,
	}
	_ = repo.UpdateConfig(ctx, restore)
}

func TestUpdateConfig_InvalidOwnerIDReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.UpdateConfig(ctx, domain.ConfigInput{OwnerId: "abc"})
	if err == nil {
		t.Fatal("expected error for non-numeric ownerID, got nil")
	}
}

func TestGetConfig_MetaTagsIsSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cfg, err := repo.GetConfig(ctx, strconv.FormatInt(testutil.OwnerID, 10))
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	// MetaTags must not be nil (may be empty).
	if cfg.MetaTags == nil {
		t.Error("expected non-nil MetaTags slice")
	}
}
