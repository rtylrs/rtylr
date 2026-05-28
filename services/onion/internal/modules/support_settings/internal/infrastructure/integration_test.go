package infrastructure_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/voxire/onion/internal/modules/support_settings/internal/domain"
	"github.com/voxire/onion/internal/modules/support_settings/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func jsonEqual(t *testing.T, want, got string) bool {
	t.Helper()
	var wm, gm interface{}
	if json.Unmarshal([]byte(want), &wm) != nil || json.Unmarshal([]byte(got), &gm) != nil {
		return want == got
	}
	return reflect.DeepEqual(wm, gm)
}

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
	if !jsonEqual(t, bh, s.BusinessHours) {
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

func TestSupportSettings_Upsert_SLAHoursZero_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	if err := repo.Upsert(context.Background(), domain.SettingsInput{
		BusinessID: testSSBusinessID,
		SLAHours:   0,
	}); err != nil {
		t.Fatalf("Upsert SLAHours=0: %v", err)
	}
}

func TestSupportSettings_Upsert_AutoAssignFalse_Persists(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	if err := repo.Upsert(context.Background(), domain.SettingsInput{
		BusinessID: testSSBusinessID,
		AutoAssign: false,
		SLAHours:   24,
	}); err != nil {
		t.Fatalf("Upsert AutoAssign=false: %v", err)
	}
	s, err := repo.Get(context.Background(), testSSBusinessID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if s.AutoAssign {
		t.Error("expected AutoAssign=false after upsert")
	}
}

func TestSupportSettings_Upsert_MultipleFields_AllPersist(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	bh := `{"tuesday":{"start":"08:00","end":"18:00"}}`
	if err := repo.Upsert(context.Background(), domain.SettingsInput{
		BusinessID:        testSSBusinessID,
		AutoAssign:        true,
		SLAHours:          36,
		NotificationEmail: "alerts@company.com",
		BusinessHours:     bh,
	}); err != nil {
		t.Fatalf("Upsert multi-field: %v", err)
	}

	s, err := repo.Get(context.Background(), testSSBusinessID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !s.AutoAssign {
		t.Error("expected AutoAssign=true")
	}
	if s.SLAHours != 36 {
		t.Errorf("expected SLAHours 36, got %d", s.SLAHours)
	}
	if s.NotificationEmail != "alerts@company.com" {
		t.Errorf("expected email alerts@company.com, got %s", s.NotificationEmail)
	}
	if !jsonEqual(t, bh, s.BusinessHours) {
		t.Errorf("expected business hours %s, got %s", bh, s.BusinessHours)
	}
}

func TestSupportSettings_Upsert_OverwritesPreviousValues(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	if err := repo.Upsert(context.Background(), domain.SettingsInput{
		BusinessID:        testSSBusinessID,
		SLAHours:          10,
		NotificationEmail: "first@example.com",
	}); err != nil {
		t.Fatalf("first Upsert: %v", err)
	}
	if err := repo.Upsert(context.Background(), domain.SettingsInput{
		BusinessID:        testSSBusinessID,
		SLAHours:          99,
		NotificationEmail: "second@example.com",
	}); err != nil {
		t.Fatalf("second Upsert: %v", err)
	}

	s, err := repo.Get(context.Background(), testSSBusinessID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if s.SLAHours != 99 {
		t.Errorf("expected SLAHours 99, got %d", s.SLAHours)
	}
	if s.NotificationEmail != "second@example.com" {
		t.Errorf("expected email second@example.com, got %s", s.NotificationEmail)
	}
}

func TestSupportSettings_Get_UnknownBusiness_ReturnsDefaults(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	s, err := repo.Get(context.Background(), 9999997)
	if err != nil {
		t.Fatalf("Get unknown business: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil settings for unknown business")
	}
}

func TestSupportSettings_Upsert_EmptyEmail_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	if err := repo.Upsert(context.Background(), domain.SettingsInput{
		BusinessID:        testSSBusinessID,
		SLAHours:          24,
		NotificationEmail: "",
	}); err != nil {
		t.Fatalf("Upsert empty email: %v", err)
	}
}

func TestSupportSettings_Upsert_LargeSLAHours(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	if err := repo.Upsert(context.Background(), domain.SettingsInput{
		BusinessID: testSSBusinessID,
		SLAHours:   720,
	}); err != nil {
		t.Fatalf("Upsert SLAHours=720: %v", err)
	}
	s, err := repo.Get(context.Background(), testSSBusinessID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if s.SLAHours != 720 {
		t.Errorf("expected SLAHours 720, got %d", s.SLAHours)
	}
}

func TestSupportSettings_Get_ReturnsNonNilStruct(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	if err := repo.Upsert(context.Background(), domain.SettingsInput{
		BusinessID: testSSBusinessID, SLAHours: 24,
	}); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	s, err := repo.Get(context.Background(), testSSBusinessID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil settings")
	}
}

func TestSupportSettings_Upsert_EmptyBusinessHours_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	if err := repo.Upsert(context.Background(), domain.SettingsInput{
		BusinessID:    testSSBusinessID,
		SLAHours:      24,
		BusinessHours: "",
	}); err != nil {
		t.Fatalf("Upsert empty BusinessHours: %v", err)
	}
}

func TestSupportSettings_Upsert_ConsecutiveCallsRetainLatest(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	for i := 1; i <= 5; i++ {
		hours := i * 12
		if err := repo.Upsert(context.Background(), domain.SettingsInput{
			BusinessID: testSSBusinessID,
			SLAHours:   hours,
		}); err != nil {
			t.Fatalf("Upsert iteration %d: %v", i, err)
		}
	}
	s, err := repo.Get(context.Background(), testSSBusinessID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if s.SLAHours != 60 {
		t.Errorf("expected SLAHours 60 after 5 upserts, got %d", s.SLAHours)
	}
}
