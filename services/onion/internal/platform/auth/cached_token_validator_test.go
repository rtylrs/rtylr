package auth

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type countingTokenValidator struct {
	calls  atomic.Int64
	claims *TokenClaims
	err    error
}

func (v *countingTokenValidator) ValidateToken(context.Context, string) (int64, error) {
	claims, err := v.ValidateTokenFull(context.Background(), "")
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

func (v *countingTokenValidator) ValidateTokenFull(context.Context, string) (*TokenClaims, error) {
	v.calls.Add(1)
	if v.err != nil {
		return nil, v.err
	}
	return cloneTokenClaims(v.claims), nil
}

func TestCachingTokenValidatorCachesValidClaims(t *testing.T) {
	inner := &countingTokenValidator{
		claims: &TokenClaims{
			UserID: 7,
			ExpiresAt: time.Now().
				Add(time.Hour),
		},
	}
	validator := NewCachingTokenValidator(inner, time.Minute)

	claims, err := validator.ValidateTokenFull(context.Background(), "token-a")
	if err != nil {
		t.Fatalf("ValidateTokenFull: %v", err)
	}
	claims.UserID = 99

	userID, err := validator.ValidateToken(context.Background(), "token-a")
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if userID != 7 {
		t.Fatalf("cached user id = %d, want 7", userID)
	}
	if calls := inner.calls.Load(); calls != 1 {
		t.Fatalf("inner calls = %d, want 1", calls)
	}
}

func TestCachingTokenValidatorDoesNotCacheErrors(t *testing.T) {
	expectedErr := errors.New("bad token")
	inner := &countingTokenValidator{
		err: expectedErr,
	}
	validator := NewCachingTokenValidator(inner, time.Minute)

	for i := 0; i < 2; i++ {
		if _, err := validator.ValidateTokenFull(context.Background(), "token-a"); !errors.Is(err, expectedErr) {
			t.Fatalf("call %d err = %v, want %v", i+1, err, expectedErr)
		}
	}
	if calls := inner.calls.Load(); calls != 2 {
		t.Fatalf("inner calls = %d, want 2", calls)
	}
}

func TestCachingTokenValidatorDoesNotCachePastTokenExpiry(t *testing.T) {
	inner := &countingTokenValidator{
		claims: &TokenClaims{
			UserID: 7,
			ExpiresAt: time.Now().
				Add(-time.Second),
		},
	}
	validator := NewCachingTokenValidator(inner, time.Minute)

	for i := 0; i < 2; i++ {
		if _, err := validator.ValidateTokenFull(context.Background(), "token-a"); err != nil {
			t.Fatalf("call %d ValidateTokenFull: %v", i+1, err)
		}
	}
	if calls := inner.calls.Load(); calls != 2 {
		t.Fatalf("inner calls = %d, want 2", calls)
	}
}

type blockingTokenValidator struct {
	calls   atomic.Int64
	started chan struct{}
	release chan struct{}
}

func (v *blockingTokenValidator) ValidateToken(context.Context, string) (int64, error) {
	claims, err := v.ValidateTokenFull(context.Background(), "")
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

func (v *blockingTokenValidator) ValidateTokenFull(context.Context, string) (*TokenClaims, error) {
	if v.calls.Add(1) == 1 {
		close(v.started)
	}
	<-v.release
	return &TokenClaims{UserID: 11, ExpiresAt: time.Now().Add(time.Hour)}, nil
}

func TestCachingTokenValidatorCollapsesConcurrentMisses(t *testing.T) {
	inner := &blockingTokenValidator{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	validator := NewCachingTokenValidator(inner, time.Minute)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			claims, err := validator.ValidateTokenFull(context.Background(), "token-a")
			if err != nil {
				t.Errorf("ValidateTokenFull: %v", err)
				return
			}
			if claims.UserID != 11 {
				t.Errorf("user id = %d, want 11", claims.UserID)
			}
		}()
	}

	<-inner.started
	close(inner.release)
	wg.Wait()

	if calls := inner.calls.Load(); calls != 1 {
		t.Fatalf("inner calls = %d, want 1", calls)
	}
}
