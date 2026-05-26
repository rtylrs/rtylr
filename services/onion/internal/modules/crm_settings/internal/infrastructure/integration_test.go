package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/crm_settings/internal/domain"
	"github.com/voxire/onion/internal/modules/crm_settings/internal/infrastructure"
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
		OwnerID:  testutil.OwnerID,
		Timezone: "Europe/London",
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
	_ = repo.Upsert(ctx, domain.SettingsInput{OwnerID: testutil.OwnerID, DefaultConsent: true})
}

func TestUpsert_SmsSenderIDPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.SettingsInput{
		OwnerID:     testutil.OwnerID,
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
		OwnerID:  testutil.OwnerID,
		FromName: "Test Sender",
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
