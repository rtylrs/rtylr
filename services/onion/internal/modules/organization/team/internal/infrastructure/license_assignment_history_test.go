package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/team/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestAuthorizationReplacementCanRevokeRepeatedLicenseHistory(t *testing.T) {
	db := testutil.OpenDB(t)
	testutil.EnsureBusiness(t, db, testutil.BusinessID, testutil.OwnerID)
	ctx := context.Background()
	result, err := db.ExecContext(ctx, `
		INSERT INTO users (business_id, email, password_hash, full_name, onboarded)
		VALUES (?, ?, 'test-password-hash', 'Authorization History Test', TRUE)`,
		testutil.BusinessID, fmt.Sprintf("authorization-history-%d@test.local", time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	userID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_audit_events WHERE target_user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_license_assignments WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_business_members WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	})

	repo := infrastructure.NewRepository(db)
	update := func(apps []string) error {
		return repo.UpsertNormalizedAuthorization(ctx, testutil.BusinessID, testutil.OwnerID, testutil.OwnerID, userID, apps, "member", "active", nil, nil)
	}
	if err := update([]string{"flow"}); err != nil {
		t.Fatalf("grant Flow: %v", err)
	}
	if err := update(nil); err != nil {
		t.Fatalf("revoke Flow: %v", err)
	}
	if err := update([]string{"flow"}); err != nil {
		t.Fatalf("regrant Flow: %v", err)
	}
	if err := update(nil); err != nil {
		t.Fatalf("revoke Flow again: %v", err)
	}

	var revoked int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM authorization_license_assignments
		WHERE business_id = ? AND user_id = ? AND app_id = 'flow' AND status = 'revoked'`,
		testutil.BusinessID, userID).Scan(&revoked); err != nil {
		t.Fatal(err)
	}
	if revoked != 2 {
		t.Fatalf("expected both revoked grants to remain in history, got %d", revoked)
	}
}
