package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/crm_settings/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/crm_settings/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGet_ReturnsDefaultsWhenNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Use an owner that doesn't have settings set up to test defaults
	settings, err := repo.Get(ctx, 8888888)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if settings.OwnerID != 8888888 {
		t.Errorf("OwnerID = %d, want 8888888", settings.OwnerID)
	}
	if settings.Timezone != "UTC" {
		t.Errorf("expected default Timezone='UTC', got %q", settings.Timezone)
	}
	if !settings.DefaultConsent {
		t.Error("expected default DefaultConsent=true")
	}
}

func TestUpsert_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.Upsert(ctx, domain.SettingsInput{
		OwnerID:        testutil.OwnerID,
		BusinessID:     testutil.BusinessID,
		DefaultConsent: true,
		FromEmail:      "noreply@example.com",
		FromName:       "Test Store",
		Timezone:       "America/New_York",
	})
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
}

func TestUpsert_ThenGet_FromEmailIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.Upsert(ctx, domain.SettingsInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FromEmail:  "stored@example.com",
		Timezone:   "UTC",
	}); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	settings, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get after upsert: %v", err)
	}
	if settings.FromEmail != "stored@example.com" {
		t.Errorf("FromEmail = %q, want %q", settings.FromEmail, "stored@example.com")
	}
}

func TestUpsert_ThenGet_TimezoneIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.Upsert(ctx, domain.SettingsInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Timezone:   "Europe/London",
	}); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	settings, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get after upsert: %v", err)
	}
	if settings.Timezone != "Europe/London" {
		t.Errorf("Timezone = %q, want %q", settings.Timezone, "Europe/London")
	}
}

func TestUpsert_ThenGet_OwnerIDIsSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.Upsert(ctx, domain.SettingsInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Timezone:   "UTC",
	}); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	settings, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if settings.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", settings.OwnerID, testutil.OwnerID)
	}
}

func TestUpsert_ThenGet_FromNameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.Upsert(ctx, domain.SettingsInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FromName:   "My Brand",
		Timezone:   "UTC",
	}); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	settings, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if settings.FromName != "My Brand" {
		t.Errorf("FromName = %q, want %q", settings.FromName, "My Brand")
	}
}

func TestUpsert_ThenGet_DefaultConsentIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.Upsert(ctx, domain.SettingsInput{
		OwnerID:        testutil.OwnerID,
		BusinessID:     testutil.BusinessID,
		DefaultConsent: false,
		Timezone:       "UTC",
	}); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	settings, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if settings.DefaultConsent != false {
		t.Error("expected DefaultConsent=false, got true")
	}
}

func TestUpsert_Idempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.SettingsInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FromEmail:  "idempotent@test.com",
		Timezone:   "UTC",
	}

	if err := repo.Upsert(ctx, input); err != nil {
		t.Fatalf("Upsert first time: %v", err)
	}
	if err := repo.Upsert(ctx, input); err != nil {
		t.Fatalf("Upsert second time: %v", err)
	}

	settings, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if settings.FromEmail != "idempotent@test.com" {
		t.Errorf("FromEmail = %q, want idempotent@test.com", settings.FromEmail)
	}
}
