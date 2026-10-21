package infrastructure_test

// Database-backed proof of the email-verification binding. Skipped unless
// DB_HOST is set, exactly like the rest of this package, so CI without a
// database does not run it. The driver-level tests in this package cover the
// same properties without one.

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/user/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func setupVerificationUser(t *testing.T, db *sql.DB) (int64, string) {
	t.Helper()

	email := fmt.Sprintf("verify+%d@example.com", time.Now().UnixNano())
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO users (email, password_hash) VALUES (?, '$2a$10$fakehash')`, email)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	var id int64
	if err := db.QueryRowContext(context.Background(),
		`SELECT id FROM users WHERE email = ?`, email).Scan(&id); err != nil {
		t.Fatalf("read seeded user id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM user_email_verification_tokens WHERE user_id = ?`, id)
		_, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE id = ?`, id)
	})

	return id, email
}

// The digest in the table is keyed to the user, so the row for one account can
// never be spent by a code the endpoint recomputed for another — and the stored
// value is not a bare digest of the code a dump could be enumerated against.
func TestEmailVerification_StoredDigestIsUserBound(t *testing.T) {
	t.Setenv("EMAIL_VERIFICATION_SECRET", "integration-secret")

	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	userID, email := setupVerificationUser(t, db)

	issued, err := repo.CreateEmailVerificationCode(context.Background(), email)
	if err != nil {
		t.Fatalf("CreateEmailVerificationCode: %v", err)
	}
	if len(issued.Code) != 10 {
		t.Fatalf("issued code %q is %d characters, want 10", issued.Code, len(issued.Code))
	}

	var stored string
	if err := db.QueryRowContext(context.Background(),
		`SELECT token_hash FROM user_email_verification_tokens WHERE user_id = ?`, userID).Scan(&stored); err != nil {
		t.Fatalf("read stored digest: %v", err)
	}
	// sha256 of the bare code, which is what the pre-binding implementation
	// stored.
	legacy := sha256Hex(issued.Code)
	if stored == legacy {
		t.Fatalf("stored digest equals the unkeyed sha256 of the code: %s", stored)
	}
	if len(stored) != 64 {
		t.Fatalf("stored digest is %d characters, want a 64-char hex sha256", len(stored))
	}
}

// A code issued for user A, presented as user B's token, must not verify
// anything — and must not touch B's provisioning state.
func TestEmailVerification_CodeDoesNotVerifyAnotherAccount(t *testing.T) {
	t.Setenv("EMAIL_VERIFICATION_SECRET", "integration-secret")

	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	ownerID, ownerEmail := setupVerificationUser(t, db)
	otherID, _ := setupVerificationUser(t, db)

	issued, err := repo.CreateEmailVerificationCode(context.Background(), ownerEmail)
	if err != nil {
		t.Fatalf("CreateEmailVerificationCode: %v", err)
	}
	if issued.UserID != ownerID {
		t.Fatalf("code issued for user %d, want %d", issued.UserID, ownerID)
	}

	// The attacker names their own user id and reuses the victim's code.
	crossToken := fmt.Sprintf("%d.%s", otherID, issued.Code)
	if _, err := repo.ConsumeEmailVerificationCode(context.Background(), crossToken); !errors.Is(err, infrastructure.ErrInvalidEmailVerificationCode) {
		t.Fatalf("cross-user token = %v, want ErrInvalidEmailVerificationCode", err)
	}

	// The owner's row is still live.
	var consumed sql.NullTime
	if err := db.QueryRowContext(context.Background(),
		`SELECT consumed_at FROM user_email_verification_tokens WHERE user_id = ? AND token_hash <> ''`,
		ownerID).Scan(&consumed); err != nil {
		t.Fatalf("read owner token row: %v", err)
	}
	if consumed.Valid {
		t.Fatal("the victim's code was spent by another account's token")
	}

	// The owner's own token works.
	if _, err := repo.ConsumeEmailVerificationCode(context.Background(), issued.Token); err != nil {
		t.Fatalf("owner's own token: %v", err)
	}
	if err := db.QueryRowContext(context.Background(),
		`SELECT consumed_at FROM user_email_verification_tokens WHERE user_id = ?`, ownerID).Scan(&consumed); err != nil {
		t.Fatalf("re-read owner token row: %v", err)
	}
	if !consumed.Valid {
		t.Fatal("the owner's own token did not spend the row")
	}
}

// A pre-binding six-digit code is refused, and a spent code cannot be replayed.
func TestEmailVerification_LegacyAndReplayAreRefused(t *testing.T) {
	t.Setenv("EMAIL_VERIFICATION_SECRET", "integration-secret")

	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	_, email := setupVerificationUser(t, db)

	if _, err := repo.ConsumeEmailVerificationCode(context.Background(), "123456"); !errors.Is(err, infrastructure.ErrSupersededEmailVerificationCode) {
		t.Fatalf("legacy code = %v, want ErrSupersededEmailVerificationCode", err)
	}

	issued, err := repo.CreateEmailVerificationCode(context.Background(), email)
	if err != nil {
		t.Fatalf("CreateEmailVerificationCode: %v", err)
	}
	if _, err := repo.ConsumeEmailVerificationCode(context.Background(), issued.Token); err != nil {
		t.Fatalf("first consume: %v", err)
	}
	if _, err := repo.ConsumeEmailVerificationCode(context.Background(), issued.Token); !errors.Is(err, infrastructure.ErrInvalidEmailVerificationCode) {
		t.Fatalf("replay = %v, want ErrInvalidEmailVerificationCode", err)
	}
}

// An HMAC key change (a JWT_SECRET rotation) invalidates outstanding codes
// rather than leaving them verifiable under a stale key.
func TestEmailVerification_SecretRotationInvalidatesOutstandingCodes(t *testing.T) {
	t.Setenv("EMAIL_VERIFICATION_SECRET", "secret-a")

	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	_, email := setupVerificationUser(t, db)

	issued, err := repo.CreateEmailVerificationCode(context.Background(), email)
	if err != nil {
		t.Fatalf("CreateEmailVerificationCode: %v", err)
	}

	t.Setenv("EMAIL_VERIFICATION_SECRET", "secret-b")
	if _, err := repo.ConsumeEmailVerificationCode(context.Background(), issued.Token); !errors.Is(err, infrastructure.ErrInvalidEmailVerificationCode) {
		t.Fatalf("after rotation = %v, want ErrInvalidEmailVerificationCode", err)
	}
}

// sha256Hex is the pre-binding stored value: an unkeyed digest of the code.
func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))

	return hex.EncodeToString(sum[:])
}
