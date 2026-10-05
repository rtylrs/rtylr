package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/crm_settings/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/crm_settings/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestUpsert_ThenGet_SmsSenderIDIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.Upsert(ctx, domain.SettingsInput{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		SmsSenderID: "MYBRAND",
		Timezone:    "UTC",
	}); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	settings, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get after upsert: %v", err)
	}
	if settings.SmsSenderID != "MYBRAND" {
		t.Errorf("SmsSenderID = %q, want %q", settings.SmsSenderID, "MYBRAND")
	}
}

func TestUpsert_ThenGet_UpdatesExistingRecord(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.Upsert(ctx, domain.SettingsInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FromEmail:  "first@example.com",
		Timezone:   "UTC",
	}); err != nil {
		t.Fatalf("Upsert first: %v", err)
	}

	if err := repo.Upsert(ctx, domain.SettingsInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FromEmail:  "second@example.com",
		Timezone:   "UTC",
	}); err != nil {
		t.Fatalf("Upsert second: %v", err)
	}

	settings, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if settings.FromEmail != "second@example.com" {
		t.Errorf("FromEmail = %q, want second@example.com", settings.FromEmail)
	}
}

func TestUpsert_DefaultConsentTrue_IsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.Upsert(ctx, domain.SettingsInput{
		OwnerID:        testutil.OwnerID,
		BusinessID:     testutil.BusinessID,
		DefaultConsent: true,
		Timezone:       "UTC",
	}); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	settings, err := repo.Get(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !settings.DefaultConsent {
		t.Error("expected DefaultConsent=true after upsert")
	}
}

func TestGet_NewOwner_TimezoneDefaultsToUTC(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	settings, err := repo.Get(ctx, 6543210)
	if err != nil {
		t.Fatalf("Get for new owner: %v", err)
	}
	if settings.Timezone != "UTC" {
		t.Errorf("expected default Timezone=UTC, got %q", settings.Timezone)
	}
}
