package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/support_settings/internal/domain"
	"github.com/voxire/onion/internal/modules/support_settings/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const testSSBusinessID int64 = 1

func TestSupportSettings_Get_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	s, err := repo.Get(context.Background(), testSSBusinessID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil settings")
	}
}

func TestSupportSettings_Get_DefaultSLAHours(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	// Get settings for a business that may not have any yet - should return defaults
	s, err := repo.Get(context.Background(), 9999998)
	if err != nil {
		t.Fatalf("Get with missing settings: %v", err)
	}
	if s.SLAHours != 24 {
		t.Errorf("expected default SLA hours 24, got %d", s.SLAHours)
	}
}

func TestSupportSettings_Upsert_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	if err := repo.Upsert(context.Background(), domain.SettingsInput{
		BusinessID: testSSBusinessID,
		AutoAssign: true,
		SLAHours:   48,
	}); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
}

func TestSupportSettings_Upsert_ChangesValues(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	if err := repo.Upsert(context.Background(), domain.SettingsInput{
		BusinessID: testSSBusinessID,
		AutoAssign: false,
		SLAHours:   72,
	}); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	s, err := repo.Get(context.Background(), testSSBusinessID)
	if err != nil {
		t.Fatalf("Get after Upsert: %v", err)
	}
	if s.SLAHours != 72 {
		t.Errorf("expected SLAHours 72, got %d", s.SLAHours)
	}
}

func TestSupportSettings_Upsert_UpdatesAutoAssign(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	if err := repo.Upsert(context.Background(), domain.SettingsInput{
		BusinessID: testSSBusinessID,
		AutoAssign: true,
		SLAHours:   24,
	}); err != nil {
		t.Fatalf("Upsert (enable auto assign): %v", err)
	}

	s, err := repo.Get(context.Background(), testSSBusinessID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !s.AutoAssign {
		t.Error("expected AutoAssign=true after upsert")
	}
}

func TestSupportSettings_Upsert_WithNotificationEmail(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	if err := repo.Upsert(context.Background(), domain.SettingsInput{
		BusinessID:        testSSBusinessID,
		SLAHours:          24,
		NotificationEmail: "support@example.com",
	}); err != nil {
		t.Fatalf("Upsert with email: %v", err)
	}

	s, err := repo.Get(context.Background(), testSSBusinessID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if s.NotificationEmail != "support@example.com" {
		t.Errorf("expected email support@example.com, got %s", s.NotificationEmail)
	}
}

func TestSupportSettings_Upsert_WithBusinessHours(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	bh := `{"monday":{"start":"09:00","end":"17:00"}}`
	if err := repo.Upsert(context.Background(), domain.SettingsInput{
		BusinessID:    testSSBusinessID,
		SLAHours:      24,
		BusinessHours: bh,
	}); err != nil {
		t.Fatalf("Upsert with business hours: %v", err)
	}

	s, err := repo.Get(context.Background(), testSSBusinessID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if s.BusinessHours != bh {
		t.Errorf("expected business hours %s, got %s", bh, s.BusinessHours)
	}
}

func TestSupportSettings_Upsert_Idempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	input := domain.SettingsInput{
		BusinessID: testSSBusinessID,
		AutoAssign: true,
		SLAHours:   12,
	}
	for i := 0; i < 3; i++ {
		if err := repo.Upsert(context.Background(), input); err != nil {
			t.Fatalf("Upsert iteration %d: %v", i, err)
		}
	}
}

func TestSupportSettings_Get_ReturnsBusinessID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	s, err := repo.Get(context.Background(), testSSBusinessID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if s.BusinessID != testSSBusinessID {
		t.Errorf("expected business_id %d, got %d", testSSBusinessID, s.BusinessID)
	}
}
