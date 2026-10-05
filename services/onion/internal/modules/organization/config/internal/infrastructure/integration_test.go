package infrastructure_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/config/internal/domain"
	"github.com/voxire/onion/internal/modules/organization/config/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
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

func TestGetThemeByOwnerID_UnknownOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	theme, err := repo.GetThemeByOwnerID(ctx, "999888777")
	if err != nil {
		t.Fatalf("GetThemeByOwnerID for unknown owner: %v", err)
	}
	if theme.ID != 0 {
		t.Errorf("expected zero ID for unknown owner, got %d", theme.ID)
	}
}

func TestUpdateConfig_ContactEmailRoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ownerStr := strconv.FormatInt(testutil.OwnerID, 10)
	original, err := repo.GetConfig(ctx, ownerStr)
	if err != nil {
		t.Fatalf("GetConfig original: %v", err)
	}

	newEmail := "integration-test@example.com"
	input := domain.ConfigInput{
		OwnerId:      ownerStr,
		SiteName:     original.SiteName,
		ContactEmail: newEmail,
	}
	if err := repo.UpdateConfig(ctx, input); err != nil {
		t.Fatalf("UpdateConfig: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.UpdateConfig(ctx, domain.ConfigInput{
			OwnerId:      ownerStr,
			SiteName:     original.SiteName,
			ContactEmail: original.ContactEmail,
		})
	})

	updated, err := repo.GetConfig(ctx, ownerStr)
	if err != nil {
		t.Fatalf("GetConfig after update: %v", err)
	}
	if updated.ContactEmail != newEmail {
		t.Errorf("ContactEmail = %q, want %q", updated.ContactEmail, newEmail)
	}
}

func TestUpdateConfig_MaintenanceModeRoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ownerStr := strconv.FormatInt(testutil.OwnerID, 10)
	original, err := repo.GetConfig(ctx, ownerStr)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}

	if err := repo.UpdateConfig(ctx, domain.ConfigInput{
		OwnerId:         ownerStr,
		SiteName:        original.SiteName,
		ContactEmail:    original.ContactEmail,
		MaintenanceMode: true,
	}); err != nil {
		t.Fatalf("UpdateConfig: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.UpdateConfig(ctx, domain.ConfigInput{
			OwnerId:         ownerStr,
			SiteName:        original.SiteName,
			ContactEmail:    original.ContactEmail,
			MaintenanceMode: original.MaintenanceMode,
		})
	})

	updated, err := repo.GetConfig(ctx, ownerStr)
	if err != nil {
		t.Fatalf("GetConfig after maintenance update: %v", err)
	}
	if !updated.MaintenanceMode {
		t.Error("expected MaintenanceMode=true after update")
	}
}

func TestGetConfig_OwnerId_IsString(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ownerStr := strconv.FormatInt(testutil.OwnerID, 10)
	cfg, err := repo.GetConfig(ctx, ownerStr)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.OwnerId == "" {
		t.Error("expected non-empty OwnerId string in config")
	}
}

func TestUpdateConfig_EnableReviewsRoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ownerStr := strconv.FormatInt(testutil.OwnerID, 10)
	original, err := repo.GetConfig(ctx, ownerStr)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}

	if err := repo.UpdateConfig(ctx, domain.ConfigInput{
		OwnerId:       ownerStr,
		SiteName:      original.SiteName,
		ContactEmail:  original.ContactEmail,
		EnableReviews: true,
	}); err != nil {
		t.Fatalf("UpdateConfig: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.UpdateConfig(ctx, domain.ConfigInput{
			OwnerId:       ownerStr,
			SiteName:      original.SiteName,
			ContactEmail:  original.ContactEmail,
			EnableReviews: original.EnableReviews,
		})
	})

	updated, err := repo.GetConfig(ctx, ownerStr)
	if err != nil {
		t.Fatalf("GetConfig after update: %v", err)
	}
	if !updated.EnableReviews {
		t.Error("expected EnableReviews=true after update")
	}
}

func TestGetConfig_SiteNameNonEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cfg, err := repo.GetConfig(ctx, strconv.FormatInt(testutil.OwnerID, 10))
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.SiteName == "" {
		t.Error("expected non-empty SiteName")
	}
}

func TestUpdateConfig_Idempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ownerStr := strconv.FormatInt(testutil.OwnerID, 10)
	original, err := repo.GetConfig(ctx, ownerStr)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}

	input := domain.ConfigInput{
		OwnerId:      ownerStr,
		SiteName:     original.SiteName,
		ContactEmail: original.ContactEmail,
	}
	if err := repo.UpdateConfig(ctx, input); err != nil {
		t.Fatalf("UpdateConfig first: %v", err)
	}
	if err := repo.UpdateConfig(ctx, input); err != nil {
		t.Fatalf("UpdateConfig second: %v", err)
	}

	after, err := repo.GetConfig(ctx, ownerStr)
	if err != nil {
		t.Fatalf("GetConfig after: %v", err)
	}
	if after.SiteName != original.SiteName {
		t.Errorf("SiteName changed unexpectedly: %q", after.SiteName)
	}
}

func TestGetConfig_FooterSubTextField(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ownerStr := strconv.FormatInt(testutil.OwnerID, 10)
	cfg, err := repo.GetConfig(ctx, ownerStr)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	_ = cfg.FooterSubText
}

func TestUpdateConfig_FooterSubTextRoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ownerStr := strconv.FormatInt(testutil.OwnerID, 10)
	original, err := repo.GetConfig(ctx, ownerStr)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}

	newFooter := "Integration Test Footer"
	if err := repo.UpdateConfig(ctx, domain.ConfigInput{
		OwnerId:       ownerStr,
		SiteName:      original.SiteName,
		ContactEmail:  original.ContactEmail,
		FooterSubText: newFooter,
	}); err != nil {
		t.Fatalf("UpdateConfig: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.UpdateConfig(ctx, domain.ConfigInput{
			OwnerId:       ownerStr,
			SiteName:      original.SiteName,
			ContactEmail:  original.ContactEmail,
			FooterSubText: original.FooterSubText,
		})
	})

	updated, err := repo.GetConfig(ctx, ownerStr)
	if err != nil {
		t.Fatalf("GetConfig after footer update: %v", err)
	}
	if updated.FooterSubText != newFooter {
		t.Errorf("FooterSubText = %q, want %q", updated.FooterSubText, newFooter)
	}
}

func TestGetConfig_BusinessSlugField(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cfg, err := repo.GetConfig(ctx, strconv.FormatInt(testutil.OwnerID, 10))
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	_ = cfg.BusinessSlug
}

func TestGetConfig_AddonsEnabledField(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cfg, err := repo.GetConfig(ctx, strconv.FormatInt(testutil.OwnerID, 10))
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	_ = cfg.AddonsEnabled
}

func TestUpdateConfig_MultipleFieldsRoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ownerStr := strconv.FormatInt(testutil.OwnerID, 10)
	original, err := repo.GetConfig(ctx, ownerStr)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}

	input := domain.ConfigInput{
		OwnerId:        ownerStr,
		SiteName:       "Multi Field Test",
		ContactEmail:   "multi@example.com",
		ContactNumber:  original.ContactNumber,
		SenderEmail:    "sender@example.com",
		Website:        "https://example.com",
		MapsURL:        "https://maps.example.com/store",
		Address:        "123 Test Street",
		City:           "Beirut",
		Zipcode:        "1100",
		Country:        "Lebanon",
		FooterSubText:  "Multi Footer",
		EnableReviews:  true,
		EnableWishlist: true,
	}
	if err := repo.UpdateConfig(ctx, input); err != nil {
		t.Fatalf("UpdateConfig: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.UpdateConfig(ctx, domain.ConfigInput{
			OwnerId:        ownerStr,
			SiteName:       original.SiteName,
			ContactEmail:   original.ContactEmail,
			ContactNumber:  original.ContactNumber,
			SenderEmail:    original.SenderEmail,
			Website:        original.Website,
			MapsURL:        original.MapsURL,
			Address:        original.Address,
			City:           original.City,
			Zipcode:        original.Zipcode,
			Country:        original.Country,
			FooterSubText:  original.FooterSubText,
			EnableReviews:  original.EnableReviews,
			EnableWishlist: original.EnableWishlist,
		})
	})

	updated, err := repo.GetConfig(ctx, ownerStr)
	if err != nil {
		t.Fatalf("GetConfig after multi-field update: %v", err)
	}
	if updated.SiteName != "Multi Field Test" {
		t.Errorf("SiteName = %q, want 'Multi Field Test'", updated.SiteName)
	}
	if updated.ContactEmail != "multi@example.com" {
		t.Errorf("ContactEmail = %q, want 'multi@example.com'", updated.ContactEmail)
	}
	if updated.Website != "https://example.com" {
		t.Errorf("Website = %q, want 'https://example.com'", updated.Website)
	}
	if updated.MapsURL != "https://maps.example.com/store" {
		t.Errorf("MapsURL = %q, want 'https://maps.example.com/store'", updated.MapsURL)
	}
	if updated.SenderEmail != "sender@example.com" {
		t.Errorf("SenderEmail = %q, want 'sender@example.com'", updated.SenderEmail)
	}
	if updated.Address != "123 Test Street" || updated.City != "Beirut" || updated.Zipcode != "1100" || updated.Country != "Lebanon" {
		t.Errorf("address fields = %q, %q, %q, %q", updated.Address, updated.City, updated.Zipcode, updated.Country)
	}
}
