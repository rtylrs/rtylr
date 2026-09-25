package auth

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type contextBlockingTokenValidator struct {
	calls   atomic.Int64
	started chan struct{}
	release chan struct{}
}

func (v *contextBlockingTokenValidator) ValidateToken(ctx context.Context, token string) (int64, error) {
	claims, err := v.ValidateTokenFull(ctx, token)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

func (v *contextBlockingTokenValidator) ValidateTokenFull(ctx context.Context, _ string) (*TokenClaims, error) {
	if v.calls.Add(1) == 1 {
		close(v.started)
	}

	select {
	case <-v.release:
		return &TokenClaims{UserID: 11, ExpiresAt: time.Now().Add(time.Hour)}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func TestCachingTokenValidatorDoesNotPoisonConcurrentWaitersWhenCallerCancels(t *testing.T) {
	inner := &contextBlockingTokenValidator{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	validator := NewCachingTokenValidator(inner, time.Minute)

	firstCtx, cancelFirst := context.WithCancel(context.Background())
	defer cancelFirst()

	firstDone := make(chan error, 1)
	go func() {
		_, err := validator.ValidateTokenFull(firstCtx, "token-a")
		firstDone <- err
	}()

	<-inner.started

	secondDone := make(chan error, 1)
	go func() {
		claims, err := validator.ValidateTokenFull(context.Background(), "token-a")
		if err != nil {
			secondDone <- err
			return
		}
		if claims.UserID != 11 {
			secondDone <- errors.New("unexpected user id")
			return
		}
		secondDone <- nil
	}()

	cancelFirst()

	select {
	case err := <-firstDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("first caller err = %v, want context canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("first caller did not return after cancellation")
	}

	close(inner.release)

	select {
	case err := <-secondDone:
		if err != nil {
			t.Fatalf("second caller err = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("second caller did not receive shared result")
	}

	if calls := inner.calls.Load(); calls != 1 {
		t.Fatalf("inner calls = %d, want 1", calls)
	}
}
