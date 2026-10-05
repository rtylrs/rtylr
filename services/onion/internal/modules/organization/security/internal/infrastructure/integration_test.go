package infrastructure_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/security/internal/infrastructure"
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

// A failure of the second write in SetTwoFactor must roll back the first, so
// users.two_factor_enabled and user_provisioning_states.mfa_required never
// disagree after a reported failure.
func TestSetTwoFactor_SecondWriteFailureRollsBackFirst(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	res, err := db.ExecContext(ctx,
		"INSERT INTO users (business_id, email, password_hash) VALUES (?, ?, '')",
		testutil.BusinessID, fmt.Sprintf("mfa-tx-%d@example.test", time.Now().UnixNano()))
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	userID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("user id: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM user_provisioning_states WHERE user_id = ?", userID)
		db.ExecContext(context.Background(), "DELETE FROM users WHERE id = ?", userID)
	})
	if _, err := db.ExecContext(ctx, `
		INSERT INTO user_provisioning_states (user_id, business_id, mode, mfa_required)
		VALUES (?, ?, 'direct_credentials', 0)`, userID, testutil.BusinessID); err != nil {
		t.Fatalf("insert provisioning state: %v", err)
	}

	// Hold the provisioning row so the second UPDATE blocks until the context
	// deadline expires.
	lockTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin lock tx: %v", err)
	}
	defer lockTx.Rollback()
	var lockedID int64
	if err := lockTx.QueryRowContext(ctx,
		`SELECT user_id FROM user_provisioning_states WHERE user_id = ? FOR UPDATE`, userID).Scan(&lockedID); err != nil {
		t.Fatalf("lock provisioning row: %v", err)
	}

	setCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	if err := repo.SetTwoFactor(setCtx, userID, true); err == nil {
		t.Fatal("SetTwoFactor reported success while the provisioning row was locked")
	}

	var enabled bool
	if err := db.QueryRowContext(ctx, `SELECT two_factor_enabled FROM users WHERE id = ?`, userID).Scan(&enabled); err != nil {
		t.Fatalf("read two_factor_enabled: %v", err)
	}
	if enabled {
		t.Fatal("users.two_factor_enabled was committed although the transaction failed")
	}
}

// insertMFATestUser creates a user with a provisioning row and returns its id.
func insertMFATestUser(t *testing.T, ctx context.Context, db *sql.DB, mfaRequired bool) int64 {
	t.Helper()
	res, err := db.ExecContext(ctx,
		"INSERT INTO users (business_id, email, password_hash) VALUES (?, ?, '')",
		testutil.BusinessID, fmt.Sprintf("mfa-rotation-%d@example.test", time.Now().UnixNano()))
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	userID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("user id: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM user_mfa_enrollments WHERE user_id = ?", userID)
		db.ExecContext(context.Background(), "DELETE FROM user_provisioning_states WHERE user_id = ?", userID)
		db.ExecContext(context.Background(), "DELETE FROM users WHERE id = ?", userID)
	})
	if _, err := db.ExecContext(ctx, `
		INSERT INTO user_provisioning_states (user_id, business_id, mode, mfa_required)
		VALUES (?, ?, 'direct_credentials', ?)`, userID, testutil.BusinessID, mfaRequired); err != nil {
		t.Fatalf("insert provisioning state: %v", err)
	}
	return userID
}

func insertTOTPRow(t *testing.T, ctx context.Context, db *sql.DB, userID int64, secret string, active bool) int64 {
	t.Helper()
	res, err := db.ExecContext(ctx, `
		INSERT INTO user_mfa_enrollments (user_id, method, secret_ciphertext, is_active)
		VALUES (?, 'totp', ?, ?)`, userID, []byte(secret), active)
	if err != nil {
		t.Fatalf("insert totp row: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("totp row id: %v", err)
	}
	return id
}

// A rotation must retire the outgoing factor. Activating the pending row alone
// left the previous TOTP secret flagged active as well, so the table held two
// live secrets and only the newest was ever consulted.
func TestCompleteMFAEnrollment_RetiresThePreviousFactor(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	userID := insertMFATestUser(t, ctx, db, true)
	outgoingID := insertTOTPRow(t, ctx, db, userID, "enc:v1:outgoing", true)
	incomingID := insertTOTPRow(t, ctx, db, userID, "enc:v1:incoming", false)

	if err := repo.CompleteMFAEnrollment(ctx, userID, [][]byte{[]byte("hash")}); err != nil {
		t.Fatalf("CompleteMFAEnrollment: %v", err)
	}

	rows, err := db.QueryContext(ctx, `
		SELECT id FROM user_mfa_enrollments
		WHERE user_id = ? AND method = 'totp' AND is_active = TRUE`, userID)
	if err != nil {
		t.Fatalf("read active factors: %v", err)
	}
	defer rows.Close()
	var active []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("scan active factor: %v", err)
		}
		active = append(active, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate active factors: %v", err)
	}
	if len(active) != 1 || active[0] != incomingID {
		t.Fatalf("active totp factors = %v, want only the incoming factor %d (outgoing was %d)", active, incomingID, outgoingID)
	}

	var revoked sql.NullTime
	if err := db.QueryRowContext(ctx,
		"SELECT revoked_at FROM user_mfa_enrollments WHERE id = ?", outgoingID).Scan(&revoked); err != nil {
		t.Fatalf("read outgoing factor: %v", err)
	}
	if !revoked.Valid {
		t.Fatal("the outgoing factor was deactivated without being recorded as revoked")
	}
}

// The step-up gate is decided from this read, so it has to distinguish "holds a
// factor" from "has proved it".
func TestGetMFAFactorState_ReportsEnrollmentAndVerification(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	userID := insertMFATestUser(t, ctx, db, true)
	state, err := repo.GetMFAFactorState(ctx, userID)
	if err != nil {
		t.Fatalf("GetMFAFactorState before enrollment: %v", err)
	}
	if state.FactorEnrolled || state.LastVerifiedAt != nil {
		t.Fatalf("state before enrollment = %+v, want no factor and no verification", state)
	}

	insertTOTPRow(t, ctx, db, userID, "enc:v1:live", true)
	state, err = repo.GetMFAFactorState(ctx, userID)
	if err != nil {
		t.Fatalf("GetMFAFactorState after enrollment: %v", err)
	}
	if !state.FactorEnrolled || state.LastVerifiedAt != nil {
		t.Fatalf("state after enrollment = %+v, want an enrolled factor and no verification", state)
	}

	if _, err := db.ExecContext(ctx, `
		UPDATE user_provisioning_states SET last_mfa_verified_at = NOW() WHERE user_id = ?`, userID); err != nil {
		t.Fatalf("stamp verification: %v", err)
	}
	state, err = repo.GetMFAFactorState(ctx, userID)
	if err != nil {
		t.Fatalf("GetMFAFactorState after verification: %v", err)
	}
	if !state.FactorEnrolled || state.LastVerifiedAt == nil {
		t.Fatalf("state after verification = %+v, want an enrolled factor and a verification time", state)
	}
}

// A user with no provisioning row at all is the one absent state that is not a
// failure: nothing has been verified, so the step-up window is simply empty.
func TestGetMFAFactorState_MissingProvisioningRowIsNotAnError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	state, err := repo.GetMFAFactorState(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetMFAFactorState(userID=0): %v", err)
	}
	if state.FactorEnrolled || state.LastVerifiedAt != nil {
		t.Fatalf("state for a non-existent user = %+v, want the zero value", state)
	}
}
