package application

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/user/internal/domain"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
	"github.com/voxire/onion/internal/testutil"
)

// fakeResetTokenStore models the cache contract the reset flow depends on:
// GetDel reads and removes an entry in one critical section, and an expired
// entry reads as absent. Get and Delete are modelled too because they are the
// shape the service must never spend a token with, and the concurrency test
// below only fails for a Get-then-Delete implementation because of them.
type fakeResetTokenStore struct {
	mu        sync.Mutex
	values    map[string][]byte
	expiresAt map[string]time.Time
	// getDelay widens the window a Get-then-Delete implementation would leave,
	// so a regression to that shape loses the race deterministically.
	getDelay time.Duration
	puts     int
	lastKey  string
	lastTTL  time.Duration
}

func newFakeResetTokenStore() *fakeResetTokenStore {
	return &fakeResetTokenStore{
		values:    make(map[string][]byte),
		expiresAt: make(map[string]time.Time),
	}
}

func (f *fakeResetTokenStore) Put(_ context.Context, key string, val []byte, ttl time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.values[key] = append([]byte(nil), val...)
	f.expiresAt[key] = time.Now().Add(ttl)
	f.puts++
	f.lastKey = key
	f.lastTTL = ttl
}

func (f *fakeResetTokenStore) Get(_ context.Context, key string) []byte {
	f.mu.Lock()
	value := f.read(key)
	delay := f.getDelay
	f.mu.Unlock()

	// A real store answers the read on this round trip and the matching delete
	// is a later one, so the delay sits outside the lock: that is the window a
	// Get-then-Delete implementation leaves for a second reader.
	if delay > 0 {
		time.Sleep(delay)
	}

	return value
}

func (f *fakeResetTokenStore) GetDel(_ context.Context, key string) []byte {
	f.mu.Lock()
	defer f.mu.Unlock()

	value := f.read(key)
	delete(f.values, key)
	delete(f.expiresAt, key)

	return value
}

func (f *fakeResetTokenStore) Delete(_ context.Context, keys []string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, key := range keys {
		delete(f.values, key)
		delete(f.expiresAt, key)
	}
}

// read must be called with the lock held.
func (f *fakeResetTokenStore) read(key string) []byte {
	value, ok := f.values[key]
	if !ok {
		return nil
	}
	if time.Now().After(f.expiresAt[key]) {
		return nil
	}

	return append([]byte(nil), value...)
}

func (f *fakeResetTokenStore) len() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return len(f.values)
}

// useResetTokenStore points the reset flow at store for the duration of the
// test. These tests must not call t.Parallel: the store is process-wide.
func useResetTokenStore(tb testing.TB, store resetTokenStore) {
	tb.Helper()

	previous := resetTokenCache
	resetTokenCache = func() resetTokenStore { return store }
	tb.Cleanup(func() { resetTokenCache = previous })
}

func seedResetToken(store *fakeResetTokenStore, token string, userID int64, ttl time.Duration) {
	store.Put(context.Background(), resetCacheKey(token), []byte(strconv.FormatInt(userID, 10)), ttl)
}

// serviceWithUnusablePasswordWrite returns a Service whose password write always
// fails. The token is spent before that write, so this covers the pre-write
// paths and the fail-closed outcome without needing a database.
func serviceWithUnusablePasswordWrite(tb testing.TB) *Service {
	tb.Helper()

	db, err := sql.Open("mysql", "reset-test:reset-test@tcp(127.0.0.1:1)/none?timeout=500ms")
	if err != nil {
		tb.Fatalf("sql.Open: %v", err)
	}
	tb.Cleanup(func() { _ = db.Close() })

	return NewService(context.Background(), db, "")
}

const testResetToken = "6f1d0c9a4b2e88715c3a90de4f6b2c81d5e7a0f39b4c6d2e8a1f0b3c5d7e9a2f"

// TestConsumeResetTokenIsSingleUse is the regression test for the replay defect:
// spending the token has to remove it in the same step that reads it.
func TestConsumeResetTokenIsSingleUse(t *testing.T) {
	ctx := context.Background()
	store := newFakeResetTokenStore()
	seedResetToken(store, testResetToken, 4242, resetTokenTTL)

	userID, err := consumeResetToken(ctx, store, testResetToken)
	if err != nil {
		t.Fatalf("first redemption: %v", err)
	}
	if userID != 4242 {
		t.Fatalf("userID = %d, want 4242", userID)
	}

	if _, err := consumeResetToken(ctx, store, testResetToken); !errors.Is(err, ErrResetTokenNotFound) {
		t.Fatalf("second redemption: got err=%v, want ErrResetTokenNotFound", err)
	}
	if store.len() != 0 {
		t.Fatalf("store still holds %d entries after redemption", store.len())
	}
}

// TestConsumeResetTokenConcurrentRedemptionHasOneWinner pins the invariant the
// non-atomic implementation broke: however many requests race for one token,
// exactly one may proceed.
func TestConsumeResetTokenConcurrentRedemptionHasOneWinner(t *testing.T) {
	const racers = 16

	ctx := context.Background()
	store := newFakeResetTokenStore()
	store.getDelay = 25 * time.Millisecond
	seedResetToken(store, testResetToken, 77, resetTokenTTL)

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		succeeded int
	)

	start := make(chan struct{})
	wg.Add(racers)
	for range racers {
		go func() {
			defer wg.Done()
			<-start

			userID, err := consumeResetToken(ctx, store, testResetToken)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil && userID == 77:
				succeeded++
			case errors.Is(err, ErrResetTokenNotFound):
			default:
				t.Errorf("unexpected redemption outcome: userID=%d err=%v", userID, err)
			}
		}()
	}
	close(start)
	wg.Wait()

	if succeeded != 1 {
		t.Fatalf("concurrent redemptions of one token: %d succeeded, want exactly 1", succeeded)
	}
}

func TestConsumeResetTokenRejectsUnusableTokens(t *testing.T) {
	ctx := context.Background()
	garbage := newFakeResetTokenStore()
	garbage.Put(ctx, resetCacheKey("garbage-token"), []byte("not-a-user-id"), resetTokenTTL)
	zeroUser := newFakeResetTokenStore()
	zeroUser.Put(ctx, resetCacheKey("zero-token"), []byte("0"), resetTokenTTL)
	expired := newFakeResetTokenStore()
	expired.Put(ctx, resetCacheKey("expired-token"), []byte("5"), -time.Minute)

	cases := []struct {
		name  string
		store resetTokenStore
		token string
		want  error
	}{
		{name: "unknown token", store: newFakeResetTokenStore(), token: testResetToken, want: ErrResetTokenNotFound},
		{name: "empty token", store: newFakeResetTokenStore(), token: "", want: ErrResetTokenNotFound},
		{name: "blank token", store: newFakeResetTokenStore(), token: "  \t ", want: ErrResetTokenNotFound},
		{name: "expired token", store: expired, token: "expired-token", want: ErrResetTokenNotFound},
		{name: "unparseable record", store: garbage, token: "garbage-token", want: ErrResetTokenInvalid},
		{name: "record for no user", store: zeroUser, token: "zero-token", want: ErrResetTokenInvalid},
		{name: "no store available", store: nil, token: testResetToken, want: ErrResetTokenNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userID, err := consumeResetToken(ctx, tc.store, tc.token)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got err=%v, want %v", err, tc.want)
			}
			if userID != 0 {
				t.Errorf("userID = %d, want 0 on a rejected token", userID)
			}
		})
	}
}

// TestResetPasswordFailedWriteStillSpendsToken covers the fail-closed half of
// the fix: once the token has been consumed, a failed password write must not
// leave it redeemable.
func TestResetPasswordFailedWriteStillSpendsToken(t *testing.T) {
	ctx := context.Background()
	store := newFakeResetTokenStore()
	useResetTokenStore(t, store)
	seedResetToken(store, testResetToken, 4242, resetTokenTTL)

	service := serviceWithUnusablePasswordWrite(t)

	err := service.ResetPassword(ctx, testResetToken, "FreshPassw0rd")
	if err == nil {
		t.Fatal("expected the unreachable password write to fail")
	}
	if errors.Is(err, ErrResetTokenNotFound) {
		t.Fatalf("first attempt reported a spent token, err=%v", err)
	}

	if _, err := consumeResetToken(ctx, store, testResetToken); !errors.Is(err, ErrResetTokenNotFound) {
		t.Fatalf("token was not spent by the failed write: err=%v", err)
	}
}

// TestResetPasswordRejectsWeakPasswordWithoutSpendingToken keeps a policy
// rejection from burning the user's only link: the check runs before the token
// is consumed and before any hashing.
func TestResetPasswordRejectsWeakPasswordWithoutSpendingToken(t *testing.T) {
	ctx := context.Background()
	store := newFakeResetTokenStore()
	useResetTokenStore(t, store)
	seedResetToken(store, testResetToken, 4242, resetTokenTTL)

	service := serviceWithUnusablePasswordWrite(t)

	// "weak" is rejected by the shared policy. The returned error must be the
	// weak-password sentinel, otherwise the API answers 500 for a 4xx condition.
	if err := service.ResetPassword(ctx, testResetToken, "weak"); !errors.Is(err, ErrWeakPassword) {
		t.Fatalf("got err=%v, want ErrWeakPassword", err)
	}

	userID, err := consumeResetToken(ctx, store, testResetToken)
	if err != nil {
		t.Fatalf("weak password must not spend the token: %v", err)
	}
	if userID != 4242 {
		t.Fatalf("userID = %d, want 4242", userID)
	}
}

func TestResetPasswordRejectsUnknownToken(t *testing.T) {
	store := newFakeResetTokenStore()
	useResetTokenStore(t, store)
	service := serviceWithUnusablePasswordWrite(t)

	if err := service.ResetPassword(context.Background(), "never-issued", "FreshPassw0rd"); !errors.Is(err, ErrResetTokenNotFound) {
		t.Fatalf("got err=%v, want ErrResetTokenNotFound", err)
	}
}

// TestResetCacheKeyHashesToken guards the keyspace: a scan or dump of the cache
// must not hand out usable reset credentials.
func TestResetCacheKeyHashesToken(t *testing.T) {
	key := resetCacheKey(testResetToken)

	if strings.Contains(key, testResetToken) {
		t.Errorf("cache key exposes the raw token: %q", key)
	}
	if !strings.HasPrefix(key, "pwd_reset:") {
		t.Errorf("cache key = %q, want a pwd_reset: prefix", key)
	}
	if _, err := hex.DecodeString(strings.TrimPrefix(key, "pwd_reset:")); err != nil {
		t.Errorf("cache key suffix is not hex: %v", err)
	}
	if len(key) != len("pwd_reset:")+64 {
		t.Errorf("cache key = %q, want a 32-byte digest", key)
	}
	if key == resetCacheKey("another-token") {
		t.Error("distinct tokens must not share a cache key")
	}
}

func TestNewResetTokenIsUnpredictable(t *testing.T) {
	first, err := newResetToken()
	if err != nil {
		t.Fatalf("newResetToken: %v", err)
	}
	second, err := newResetToken()
	if err != nil {
		t.Fatalf("newResetToken: %v", err)
	}

	if first == second {
		t.Fatal("two issued reset tokens must differ")
	}
	if len(first) != 64 {
		t.Fatalf("token length = %d, want 64 hex chars (32 random bytes)", len(first))
	}
	if _, err := hex.DecodeString(first); err != nil {
		t.Fatalf("token is not hex: %v", err)
	}
}

// TestRequestPasswordResetStoresOneHourHashedToken covers issuance against a
// real user row: one freshly minted token, under a hashed key, held for exactly
// resetTokenTTL, and nothing minted for an address with no account.
func TestRequestPasswordResetStoresOneHourHashedToken(t *testing.T) {
	db := testutil.OpenDB(t)

	ctx := context.Background()
	email := fmt.Sprintf("reset-issue+%d@example.com", time.Now().UnixNano())
	service := NewService(ctx, db, "")

	userID, err := service.repo.CreateUser(ctx, &domain.User{
		Email:        email,
		PasswordHash: "hash",
		FullName:     "Reset Issue",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM email_jobs WHERE to_email = ?", email)
		_ = service.repo.DeleteUser(context.Background(), userID)
	})

	store := newFakeResetTokenStore()
	useResetTokenStore(t, store)

	// The mail enqueue needs the email queue tables and is not what this
	// asserts; the token is already stored by the time it runs.
	if err := service.RequestPasswordReset(ctx, email); err != nil {
		t.Logf("RequestPasswordReset enqueue: %v", err)
	}

	if store.puts != 1 {
		t.Fatalf("stored %d tokens, want exactly 1", store.puts)
	}
	if store.lastTTL != resetTokenTTL {
		t.Errorf("token ttl = %v, want %v", store.lastTTL, resetTokenTTL)
	}
	if !strings.HasPrefix(store.lastKey, "pwd_reset:") || len(store.lastKey) != len("pwd_reset:")+64 {
		t.Errorf("token key = %q, want a hashed pwd_reset: key", store.lastKey)
	}
	if value := store.Get(ctx, store.lastKey); string(value) != strconv.FormatInt(userID, 10) {
		t.Errorf("token record = %q, want the issuing user id %d", value, userID)
	}

	// An address with no account must not mint anything, so the request path
	// stays indistinguishable for a valid and an unknown address.
	if err := service.RequestPasswordReset(ctx, "unknown-"+email); err != nil {
		t.Fatalf("RequestPasswordReset for an unknown address: %v", err)
	}
	if store.puts != 1 {
		t.Fatalf("a request for an unknown address minted a token (%d stored)", store.puts)
	}
}

// TestResetPasswordEndToEndRedeemsOnce exercises the whole path against a real
// user row: the password is written on the winning redemption, the loser is
// rejected, and the token is not left behind.
func TestResetPasswordEndToEndRedeemsOnce(t *testing.T) {
	db := testutil.OpenDB(t)

	ctx := context.Background()
	email := fmt.Sprintf("reset-replay+%d@example.com", time.Now().UnixNano())
	oldHash, err := passwordutils.HashPassword("OldPassw0rd")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	service := NewService(ctx, db, "")
	userID, err := service.repo.CreateUser(ctx, &domain.User{
		Email:        email,
		PasswordHash: oldHash,
		FullName:     "Reset Replay",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() { _ = service.repo.DeleteUser(context.Background(), userID) })

	store := newFakeResetTokenStore()
	useResetTokenStore(t, store)
	seedResetToken(store, testResetToken, userID, resetTokenTTL)

	const newPassword = "BrandNewPassw0rd"
	if err := service.ResetPassword(ctx, testResetToken, newPassword); err != nil {
		t.Fatalf("first redemption: %v", err)
	}

	updated, err := service.repo.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if !passwordutils.CheckPasswordHash(newPassword, updated.PasswordHash) {
		t.Fatal("the winning redemption did not write the new password")
	}

	if err := service.ResetPassword(ctx, testResetToken, "AttackerPassw0rd"); !errors.Is(err, ErrResetTokenNotFound) {
		t.Fatalf("replayed token: got err=%v, want ErrResetTokenNotFound", err)
	}

	unchanged, err := service.repo.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if !passwordutils.CheckPasswordHash(newPassword, unchanged.PasswordHash) {
		t.Fatal("a replayed token changed the password a second time")
	}
}
