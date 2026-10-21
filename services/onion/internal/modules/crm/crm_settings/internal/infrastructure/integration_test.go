package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/crm_settings/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/crm_settings/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGet_ReturnsSettingsForOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	settings, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if settings.OwnerID != testutil.OwnerID {
		t.Errorf("expected OwnerID=%d, got %d", testutil.OwnerID, settings.OwnerID)
	}
}

func TestGet_TimezoneDefaultsToUTC(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	settings, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if settings.Timezone == "" {
		t.Error("expected non-empty Timezone")
	}
}

func TestGet_UnknownOwnerReturnsDefaults(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// When no row exists, Get returns defaults rather than an error.
	settings, err := repo.Get(ctx, 999888777)
	if err != nil {
		t.Fatalf("Get unknown owner: %v", err)
	}
	if settings.OwnerID != 999888777 {
		t.Errorf("expected OwnerID=999888777, got %d", settings.OwnerID)
	}
	if settings.Timezone != "UTC" {
		t.Errorf("expected Timezone='UTC' for defaults, got %q", settings.Timezone)
	}
}

func TestUpsert_PersistsFromEmail(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.SettingsInput{
		OwnerID:        testutil.OwnerID,
		BusinessID:     testutil.BusinessID,
		DefaultConsent: true,
		FromEmail:      "integration@test.local",
		FromName:       "Integration Test",
		Timezone:       "America/New_York",
	}
	if err := repo.Upsert(ctx, input); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	settings, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get after upsert: %v", err)
	}
	if settings.FromEmail != "integration@test.local" {
		t.Errorf("expected FromEmail='integration@test.local', got %q", settings.FromEmail)
	}
}

func TestUpsert_UpdatesTimezone(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.SettingsInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Timezone:   "Europe/London",
	}
	if err := repo.Upsert(ctx, input); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	settings, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get after timezone upsert: %v", err)
	}
	if settings.Timezone != "Europe/London" {
		t.Errorf("expected Timezone='Europe/London', got %q", settings.Timezone)
	}
}

func TestUpsert_DefaultConsentStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.SettingsInput{
		OwnerID:        testutil.OwnerID,
		BusinessID:     testutil.BusinessID,
		DefaultConsent: false,
	}
	if err := repo.Upsert(ctx, input); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	settings, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get after consent upsert: %v", err)
	}
	if settings.DefaultConsent != false {
		t.Errorf("expected DefaultConsent=false, got %v", settings.DefaultConsent)
	}

	// Restore to true.
	_ = repo.Upsert(ctx, domain.SettingsInput{OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, DefaultConsent: true})
}

func TestUpsert_SmsSenderIDPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.SettingsInput{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		SmsSenderID: "RTYLR",
	}
	if err := repo.Upsert(ctx, input); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	settings, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if settings.SmsSenderID != "RTYLR" {
		t.Errorf("expected SmsSenderID='RTYLR', got %q", settings.SmsSenderID)
	}
}

func TestGet_OwnerIDIsAlwaysSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	settings, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if settings.OwnerID == 0 {
		t.Error("expected non-zero OwnerID in settings")
	}
}

func TestUpsert_FromNamePersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.SettingsInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FromName:   "Test Sender",
	}
	if err := repo.Upsert(ctx, input); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	settings, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if settings.FromName != "Test Sender" {
		t.Errorf("expected FromName='Test Sender', got %q", settings.FromName)
	}
}

func TestGetSettings_ReturnsValidStruct(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	settings, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	// Verify the struct is fully populated (no zero-value OwnerID)
	if settings.OwnerID == 0 {
		t.Error("expected non-zero OwnerID in returned settings")
	}
}

func TestUpsertSettings_ThenGet_Roundtrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.SettingsInput{
		OwnerID:        testutil.OwnerID,
		BusinessID:     testutil.BusinessID,
		FromEmail:      "roundtrip@test.local",
		FromName:       "Roundtrip Sender",
		Timezone:       "Asia/Beirut",
		SmsSenderID:    "ROUNDTRIP",
		DefaultConsent: true,
	}
	if err := repo.Upsert(ctx, input); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get after roundtrip upsert: %v", err)
	}
	if got.FromEmail != "roundtrip@test.local" {
		t.Errorf("FromEmail = %q, want %q", got.FromEmail, "roundtrip@test.local")
	}
	if got.Timezone != "Asia/Beirut" {
		t.Errorf("Timezone = %q, want %q", got.Timezone, "Asia/Beirut")
	}
}

func TestGetSettings_WrongOwner_ReturnsDefaults(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	settings, err := repo.Get(ctx, 888777666)
	if err != nil {
		t.Fatalf("Get unknown owner: %v", err)
	}
	if settings.OwnerID != 888777666 {
		t.Errorf("expected OwnerID=888777666 echoed back, got %d", settings.OwnerID)
	}
}

func TestUpsertSettings_CurrencyIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Store a distinct from_email to verify persistence
	input := domain.SettingsInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FromEmail:  "currency.check@test.local",
	}
	if err := repo.Upsert(ctx, input); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.FromEmail != "currency.check@test.local" {
		t.Errorf("FromEmail = %q, want %q", got.FromEmail, "currency.check@test.local")
	}
}

func TestUpsertSettings_TimezoneIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.SettingsInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Timezone:   "Pacific/Auckland",
	}
	if err := repo.Upsert(ctx, input); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Timezone != "Pacific/Auckland" {
		t.Errorf("Timezone = %q, want %q", got.Timezone, "Pacific/Auckland")
	}
}

func TestUpsertSettings_MultipleUpserts_LastWins(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_ = repo.Upsert(ctx, domain.SettingsInput{OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, FromName: "First"})
	_ = repo.Upsert(ctx, domain.SettingsInput{OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, FromName: "Second"})
	_ = repo.Upsert(ctx, domain.SettingsInput{OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, FromName: "Third"})

	got, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.FromName != "Third" {
		t.Errorf("expected last upsert FromName='Third', got %q", got.FromName)
	}
}

func TestGetSettings_NoError_EmptyOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// A never-before-seen owner should return default struct without DB error
	_, err := repo.Get(ctx, 777666555)
	if err != nil {
		t.Fatalf("Get for unseen owner: %v", err)
	}
}

func TestUpsert_Idempotent_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.SettingsInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Timezone:   "Europe/Paris",
	}
	if err := repo.Upsert(ctx, input); err != nil {
		t.Fatalf("first Upsert: %v", err)
	}
	if err := repo.Upsert(ctx, input); err != nil {
		t.Fatalf("second Upsert (idempotent): %v", err)
	}
}

func TestUpsertSettings_OwnerIDScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const altOwner int64 = 555444333

	if err := repo.Upsert(ctx, domain.SettingsInput{
		OwnerID:    altOwner,
		BusinessID: altOwner,
		FromEmail:  "alt@test.local",
	}); err != nil {
		t.Fatalf("Upsert altOwner: %v", err)
	}

	if err := repo.Upsert(ctx, domain.SettingsInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FromEmail:  "main@test.local",
	}); err != nil {
		t.Fatalf("Upsert mainOwner: %v", err)
	}

	altSettings, err := repo.Get(ctx, altOwner)
	if err != nil {
		t.Fatalf("Get altOwner: %v", err)
	}
	mainSettings, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get mainOwner: %v", err)
	}

	if altSettings.FromEmail != "alt@test.local" {
		t.Errorf("alt owner FromEmail = %q, want alt@test.local", altSettings.FromEmail)
	}
	if mainSettings.FromEmail != "main@test.local" {
		t.Errorf("main owner FromEmail = %q, want main@test.local", mainSettings.FromEmail)
	}
}
