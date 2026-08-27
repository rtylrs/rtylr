package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/security/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetSecuritySettings_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	settings, err := repo.GetSecuritySettings(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("security.GetSecuritySettings: unexpected error: %v", err)
	}
	// ActiveSessions must be non-nil (empty slice, not nil)
	if settings.ActiveSessions == nil {
		t.Fatal("GetSecuritySettings: ActiveSessions should not be nil")
	}
}

func TestGetSecuritySettings_NonExistentUser_ReturnsDefaults(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// userID 0 does not exist -- repo should return safe defaults, not an error
	settings, err := repo.GetSecuritySettings(ctx, 0)
	if err != nil {
		t.Fatalf("GetSecuritySettings(userID=0): expected default, got error: %v", err)
	}
	if settings.TwoFactorEnabled {
		t.Error("non-existent user should have TwoFactorEnabled=false")
	}
}

func TestGetSecuritySettings_KnownUser1001_ReturnsValidStruct(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	settings, err := repo.GetSecuritySettings(ctx, 1001)
	if err != nil {
		t.Fatalf("GetSecuritySettings(userID=1001): unexpected error: %v", err)
	}
	// The fixture user exists; we only check the struct is well-formed.
	_ = settings.TwoFactorEnabled
}

func TestGetSecuritySettings_KnownUser1002_ReturnsValidStruct(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	settings, err := repo.GetSecuritySettings(ctx, 1002)
	if err != nil {
		t.Fatalf("GetSecuritySettings(userID=1002): unexpected error: %v", err)
	}
	_ = settings.TwoFactorEnabled
}

func TestGetLoginHistory_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	history, _, err := repo.GetLoginHistory(ctx, testutil.OwnerID, 20, 0)
	if err != nil {
		t.Fatalf("security.GetLoginHistory: unexpected error: %v", err)
	}
	// May be empty; just verify no error and valid slice.
	_ = history
}

func TestGetLoginHistory_NoUser_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	history, _, err := repo.GetLoginHistory(ctx, 0, 20, 0)
	if err != nil {
		t.Fatalf("GetLoginHistory(userID=0): unexpected error: %v", err)
	}
	if len(history) != 0 {
		t.Fatalf("expected 0 entries for non-existent user, got %d", len(history))
	}
}

func TestRecordLogin_IncreasesHistoryCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Remove all login history entries for this owner to start fresh.
	db.ExecContext(ctx, "DELETE FROM login_history WHERE user_id = ?", testutil.OwnerID)
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM login_history WHERE user_id = ?", testutil.OwnerID)
	})

	before, _, err := repo.GetLoginHistory(ctx, testutil.OwnerID, 20, 0)
	if err != nil {
		t.Fatalf("GetLoginHistory before: %v", err)
	}

	err = repo.RecordLogin(ctx, testutil.OwnerID, "test@example.com", "password", "127.0.0.1", "TestAgent", true)
	if err != nil {
		t.Fatalf("RecordLogin: %v", err)
	}

	after, _, err := repo.GetLoginHistory(ctx, testutil.OwnerID, 20, 0)
	if err != nil {
		t.Fatalf("GetLoginHistory after: %v", err)
	}

	if len(after) <= len(before) {
		t.Fatalf("expected history count to increase: before=%d after=%d", len(before), len(after))
	}
}

func TestRecordLogin_FieldsPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const email = "persist@example.com"
	const method = "password"
	const ip = "10.0.0.1"
	const agent = "IntegrationTest/1.0"

	err := repo.RecordLogin(ctx, testutil.OwnerID, email, method, ip, agent, true)
	if err != nil {
		t.Fatalf("RecordLogin: %v", err)
	}

	history, _, err := repo.GetLoginHistory(ctx, testutil.OwnerID, 20, 0)
	if err != nil {
		t.Fatalf("GetLoginHistory: %v", err)
	}

	var found bool
	for _, e := range history {
		if e.Email == email && e.Method == method {
			found = true
			if !e.Success {
				t.Errorf("expected Success=true, got false")
			}
			if e.UserAgent != agent {
				t.Errorf("user agent = %q, want %q", e.UserAgent, agent)
			}
			break
		}
	}
	if !found {
		t.Fatal("recorded login entry not found in history")
	}
}

func TestRecordLogin_FailedLogin(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.RecordLogin(ctx, testutil.OwnerID, "failed@example.com", "password", "192.168.1.1", "BadAgent", false)
	if err != nil {
		t.Fatalf("RecordLogin(success=false): %v", err)
	}

	history, _, err := repo.GetLoginHistory(ctx, testutil.OwnerID, 20, 0)
	if err != nil {
		t.Fatalf("GetLoginHistory: %v", err)
	}

	var found bool
	for _, e := range history {
		if e.Email == "failed@example.com" && !e.Success {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("failed login entry not found in history")
	}
}

func TestGetLoginHistory_RespectsPageSizeAndTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Seed 25 entries
	for i := 0; i < 25; i++ {
		_ = repo.RecordLogin(ctx, testutil.OwnerID, "bulk@example.com", "password", "1.1.1.1", "BulkAgent", true)
	}

	history, total, err := repo.GetLoginHistory(ctx, testutil.OwnerID, 20, 0)
	if err != nil {
		t.Fatalf("GetLoginHistory: %v", err)
	}

	if len(history) != 20 {
		t.Errorf("expected exactly 20 entries, got %d", len(history))
	}
	if total < 25 {
		t.Errorf("expected total to include all seeded entries, got %d", total)
	}
}

func TestGetLoginHistory_EachEntryHasEmail(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.RecordLogin(ctx, testutil.OwnerID, "email_check@example.com", "password", "1.2.3.4", "UA", true)
	if err != nil {
		t.Fatalf("RecordLogin: %v", err)
	}

	history, _, err := repo.GetLoginHistory(ctx, testutil.OwnerID, 20, 0)
	if err != nil {
		t.Fatalf("GetLoginHistory: %v", err)
	}

	for _, e := range history {
		if e.Email == "" {
			t.Error("login history entry has empty Email")
		}
	}
}

func TestSetTwoFactor_Persists(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.SetTwoFactor(ctx, testutil.OwnerID, true)
	if err != nil {
		t.Fatalf("SetTwoFactor: %v", err)
	}
	settings, err := repo.GetSecuritySettings(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetSecuritySettings after SetTwoFactor: %v", err)
	}
	if !settings.TwoFactorEnabled {
		t.Fatal("expected two-factor authentication to be enabled")
	}
}
