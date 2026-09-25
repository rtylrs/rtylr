package application

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/testutil"
)

const acceptInvitePassword = "Str0ng!Passw0rd"

// ghostOwnerID returns an owner id that has no businesses row, which is the
// sql.ErrNoRows case of GetBusinessIDByOwnerID.
func ghostOwnerID(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	ownerID := time.Now().UnixNano() % 100000000
	var count int
	if err := db.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM businesses WHERE owner_id = ?", ownerID).Scan(&count); err != nil {
		t.Fatalf("count businesses for ghost owner: %v", err)
	}
	if count != 0 {
		t.Fatalf("ghost owner %d unexpectedly owns %d businesses", ownerID, count)
	}
	return ownerID
}

// insertInvitee creates a bare invited user (no password, no permissions) with
// the given business_id (nil for NULL) and removes it on cleanup.
func insertInvitee(t *testing.T, db *sql.DB, businessID *int64) int64 {
	t.Helper()
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	email := fmt.Sprintf("accept-invite-%d@test.local", suffix)
	res, err := db.ExecContext(ctx,
		`INSERT INTO users (email, username, password_hash, business_id) VALUES (?, ?, '', ?)`,
		email, fmt.Sprintf("accept-invite-%d", suffix), businessID)
	if err != nil {
		t.Fatalf("insert invitee: %v", err)
	}
	userID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM user_provisioning_states WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_business_members WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM invite_tokens WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	})
	return userID
}

func createInviteTokenFor(t *testing.T, db *sql.DB, userID, ownerID int64) string {
	t.Helper()
	token := fmt.Sprintf("accept-invite-token-%d", time.Now().UnixNano())
	if _, err := db.ExecContext(context.Background(),
		`INSERT INTO invite_tokens (user_id, owner_id, token, expires_at) VALUES (?, ?, ?, ?)`,
		userID, ownerID, token, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("insert invite token: %v", err)
	}
	return token
}

// When the business cannot be resolved, AcceptInvite must not report success:
// the old code silently skipped activation, answered 200, and left the invitee
// stuck in pending_invitation with a consumed token.
func TestAcceptInviteFailsWhenBusinessCannotBeResolved(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	service := NewService(ctx, db)

	userID := insertInvitee(t, db, nil)
	token := createInviteTokenFor(t, db, userID, ghostOwnerID(t, db))

	if _, err := service.AcceptInvite(ctx, token, acceptInvitePassword); err == nil {
		t.Fatal("AcceptInvite succeeded with no resolvable business, want error")
	}

	// The claim must be released so the invitee can retry after the failure.
	if _, _, _, err := service.repo.GetInviteToken(ctx, token); err != nil {
		t.Fatalf("invite token is no longer usable after failed accept: %v", err)
	}
}

// The member's users.business_id is the fallback when the owner lookup fails,
// so a valid invite still activates instead of silently doing nothing.
func TestAcceptInviteActivatesViaMemberBusinessWhenOwnerLookupFails(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	service := NewService(ctx, db)

	businessID := testutil.BusinessID
	// The shared fixture business can be dropped by another suite's cleanup, so
	// re-seed it (without deleting it) before writing rows that reference it.
	testutil.EnsureBusiness(t, db, businessID, testutil.OwnerID)
	userID := insertInvitee(t, db, &businessID)
	token := createInviteTokenFor(t, db, userID, ghostOwnerID(t, db))

	if _, err := db.ExecContext(ctx,
		`INSERT INTO authorization_business_members (business_id, user_id, status) VALUES (?, ?, 'pending')`,
		businessID, userID); err != nil {
		t.Fatalf("insert pending membership: %v", err)
	}

	if _, err := service.AcceptInvite(ctx, token, acceptInvitePassword); err != nil {
		t.Fatalf("AcceptInvite: %v", err)
	}

	status, found, err := service.repo.GetProvisioningStatus(ctx, userID)
	if err != nil {
		t.Fatalf("GetProvisioningStatus: %v", err)
	}
	if !found || status == nil {
		t.Fatal("provisioning state missing after accept")
	}
	if string(status.State) != "active" || !status.EmailVerified {
		t.Fatalf("provisioning state = %q email_verified=%t, want active/true",
			status.State, status.EmailVerified)
	}

	var membershipStatus string
	if err := db.QueryRowContext(ctx,
		`SELECT status FROM authorization_business_members WHERE business_id = ? AND user_id = ?`,
		businessID, userID).Scan(&membershipStatus); err != nil {
		t.Fatalf("read membership status: %v", err)
	}
	if membershipStatus != "active" {
		t.Fatalf("membership status = %q, want active", membershipStatus)
	}
}
