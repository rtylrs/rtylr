package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos/users/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/users/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const (
	attemptsBusinessID int64 = 1001
	attemptsLocationID int64 = 3001
	attemptsUserID     int64 = 1002
)

// A fresh store must not report a backoff: an operator who has not failed yet
// must be served immediately.
func TestPINAttempts_FreshKeyHasNoBackoff(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	store := infrastructure.NewPINAttemptStore(db)

	retryAt, err := store.PinRetryAt(ctx, attemptsBusinessID, attemptsLocationID, attemptsUserID)
	if err != nil {
		t.Fatalf("PinRetryAt: %v", err)
	}
	if !retryAt.IsZero() {
		t.Errorf("retryAt = %s, want the zero time for a key with no failures", retryAt)
	}
	t.Cleanup(func() { _ = store.ResetPINFailures(ctx, attemptsBusinessID, attemptsLocationID, attemptsUserID) })
}

// The counter is durable: a second store instance over the same database — which
// is what a restarted process or a second replica is — still sees the failures.
func TestPINAttempts_CounterSurvivesANewStoreInstance(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	store := infrastructure.NewPINAttemptStore(db)
	t.Cleanup(func() { _ = store.ResetPINFailures(ctx, attemptsBusinessID, attemptsLocationID, attemptsUserID) })
	if err := store.ResetPINFailures(ctx, attemptsBusinessID, attemptsLocationID, attemptsUserID); err != nil {
		t.Fatalf("seed reset: %v", err)
	}

	for i := 0; i < domain.PinVerifyMaxFailures; i++ {
		if _, err := store.RecordPINFailure(ctx, attemptsBusinessID, attemptsLocationID, attemptsUserID); err != nil {
			t.Fatalf("RecordPINFailure %d: %v", i, err)
		}
	}

	restarted := infrastructure.NewPINAttemptStore(db)
	retryAt, err := restarted.PinRetryAt(ctx, attemptsBusinessID, attemptsLocationID, attemptsUserID)
	if err != nil {
		t.Fatalf("PinRetryAt after restart: %v", err)
	}
	if !time.Now().Before(retryAt) {
		t.Errorf("retryAt = %s is not in the future after a restart: the lockout did not survive", retryAt)
	}
	if remaining := domain.PinLockoutDuration - time.Until(retryAt); remaining > time.Minute {
		t.Errorf("retryAt is %s beyond now, want the %s lockout: the backoff schedule is not being applied",
			remaining, domain.PinLockoutDuration)
	}
}

// The counter must be keyed by (business, location, user). Collapsing the key
// would let one attacker lock out a whole store.
func TestPINAttempts_KeyIsScopedToBusinessLocationAndUser(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	store := infrastructure.NewPINAttemptStore(db)
	t.Cleanup(func() {
		for _, key := range [][2]int64{
			{attemptsLocationID, attemptsUserID},
			{attemptsLocationID, 1003},
			{3002, attemptsUserID},
		} {
			_ = store.ResetPINFailures(ctx, attemptsBusinessID, key[0], key[1])
		}
		_ = store.ResetPINFailures(ctx, 999999, attemptsLocationID, attemptsUserID)
	})

	for i := 0; i < domain.PinVerifyMaxFailures; i++ {
		if _, err := store.RecordPINFailure(ctx, attemptsBusinessID, attemptsLocationID, attemptsUserID); err != nil {
			t.Fatalf("RecordPINFailure: %v", err)
		}
	}

	others := [][3]int64{
		{attemptsBusinessID, attemptsLocationID, 1003}, // different user
		{attemptsBusinessID, 3002, attemptsUserID},     // different location
		{999999, attemptsLocationID, attemptsUserID},   // different business
	}
	for _, key := range others {
		retryAt, err := store.PinRetryAt(ctx, key[0], key[1], key[2])
		if err != nil {
			t.Fatalf("PinRetryAt(%v): %v", key, err)
		}
		if !retryAt.IsZero() {
			t.Errorf("key %v inherited a backoff of %s from another operator's failures", key, retryAt)
		}
	}
}

// Repeated failures must not accumulate without bound; the schedule has to keep
// producing the full lockout rather than a number that grows forever.
func TestPINAttempts_BackoffIsCappedAtTheLockout(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	store := infrastructure.NewPINAttemptStore(db)
	t.Cleanup(func() { _ = store.ResetPINFailures(ctx, attemptsBusinessID, attemptsLocationID, attemptsUserID) })

	for i := 1; i <= domain.PinVerifyMaxFailures+20; i++ {
		retryAt, err := store.RecordPINFailure(ctx, attemptsBusinessID, attemptsLocationID, attemptsUserID)
		if err != nil {
			t.Fatalf("RecordPINFailure %d: %v", i, err)
		}
		if i < domain.PinVerifyMaxFailures {
			continue
		}
		if wait := time.Until(retryAt); wait > domain.PinLockoutDuration+time.Minute {
			t.Errorf("failure %d imposed a %s wait, want at most the %s lockout", i, wait, domain.PinLockoutDuration)
		}
	}
}

// The reset used to unstick a locked-out operator must actually clear the row,
// not merely zero a column that a later read ignores.
func TestPINAttempts_ResetClearsTheKey(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	store := infrastructure.NewPINAttemptStore(db)
	t.Cleanup(func() { _ = store.ResetPINFailures(ctx, attemptsBusinessID, attemptsLocationID, attemptsUserID) })

	for i := 0; i < domain.PinVerifyMaxFailures; i++ {
		if _, err := store.RecordPINFailure(ctx, attemptsBusinessID, attemptsLocationID, attemptsUserID); err != nil {
			t.Fatalf("RecordPINFailure: %v", err)
		}
	}
	if err := store.ResetPINFailures(ctx, attemptsBusinessID, attemptsLocationID, attemptsUserID); err != nil {
		t.Fatalf("ResetPINFailures: %v", err)
	}
	retryAt, err := store.PinRetryAt(ctx, attemptsBusinessID, attemptsLocationID, attemptsUserID)
	if err != nil {
		t.Fatalf("PinRetryAt: %v", err)
	}
	if !retryAt.IsZero() {
		t.Errorf("retryAt = %s after a reset, want the zero time", retryAt)
	}
}
