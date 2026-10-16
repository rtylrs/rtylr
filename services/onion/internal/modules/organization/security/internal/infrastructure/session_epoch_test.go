package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/security/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestInvalidateAllSessionsAlsoRevokesIssuedAccessTokens(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	res, err := db.ExecContext(ctx, `INSERT INTO users (email, password_hash) VALUES (?, 'x')`, fmt.Sprintf("epoch-%d@example.com", time.Now().UnixNano()))
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	userID, _ := res.LastInsertId()
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM user_sessions WHERE user_id = ?`, userID)
		_, _ = db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, userID)
	})
	if _, err := db.ExecContext(ctx, `INSERT INTO user_sessions (user_id, token_hash, expires_at) VALUES (?, ?, ?)`, userID, fmt.Sprintf("hash-%d", userID), time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("insert session: %v", err)
	}

	var before, after int64
	if err := db.QueryRowContext(ctx, `SELECT token_epoch FROM users WHERE id = ?`, userID).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if err := repo.InvalidateAllSessions(ctx, userID); err != nil {
		t.Fatalf("InvalidateAllSessions: %v", err)
	}
	if err := db.QueryRowContext(ctx, `SELECT token_epoch FROM users WHERE id = ?`, userID).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after != before+1 {
		t.Fatalf("token_epoch %d -> %d, want +1 so outstanding access tokens stop validating", before, after)
	}
	var live int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_sessions WHERE user_id = ? AND revoked_at IS NULL`, userID).Scan(&live); err != nil || live != 0 {
		t.Fatalf("live sessions = %d, err %v", live, err)
	}
}

func TestConsumeTOTPCounterOnlyMovesForward(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	userID := time.Now().UnixNano()
	if _, err := db.ExecContext(ctx, `INSERT INTO user_mfa_enrollments (user_id, method, secret_ciphertext, is_active) VALUES (?, 'totp', 'x', TRUE)`, userID); err != nil {
		t.Fatalf("insert factor: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, `DELETE FROM user_mfa_enrollments WHERE user_id = ?`, userID) })

	for _, step := range []struct {
		counter int64
		want    bool
	}{{100, true}, {100, false}, {99, false}, {101, true}} {
		got, err := repo.ConsumeTOTPCounter(ctx, userID, step.counter)
		if err != nil || got != step.want {
			t.Fatalf("ConsumeTOTPCounter(%d) = %v, %v; want %v", step.counter, got, err, step.want)
		}
	}
}
