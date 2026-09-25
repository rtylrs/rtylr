package api

import (
	"context"
	"testing"
	"time"
)

type fakeLoginRecorder struct {
	recorded chan error
}

func (f *fakeLoginRecorder) RecordLogin(ctx context.Context, userID int64, email, method, ip, userAgent string, success bool) error {
	if userID != 42 || email != "alice@example.com" || method != "email" || ip != "127.0.0.1" || userAgent != "TestAgent/1.0" || !success {
		f.recorded <- context.Canceled
		return nil
	}
	f.recorded <- ctx.Err()
	return nil
}

func TestRunAsyncUserHookIgnoresParentCancellation(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	cancelParent()

	done := make(chan error, 1)
	runAsyncUserHook(parent, 42, func(ctx context.Context, userID int64) {
		if userID != 42 {
			done <- context.Canceled
			return
		}
		select {
		case <-ctx.Done():
			done <- ctx.Err()
		default:
			done <- nil
		}
	})

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected hook context to survive parent cancellation, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for async hook")
	}
}

func TestRunAsyncLoginTasksIgnoresParentCancellation(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	cancelParent()

	recorder := &fakeLoginRecorder{recorded: make(chan error, 1)}
	hookDone := make(chan error, 1)
	runAsyncLoginTasks(parent, recorder, 42, "alice@example.com", "127.0.0.1", "TestAgent/1.0", func(ctx context.Context, userID int64) {
		if userID != 42 {
			hookDone <- context.Canceled
			return
		}
		hookDone <- ctx.Err()
	})

	select {
	case err := <-recorder.recorded:
		if err != nil {
			t.Fatalf("expected login record context to survive parent cancellation, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for login recorder")
	}

	select {
	case err := <-hookDone:
		if err != nil {
			t.Fatalf("expected login hook context to survive parent cancellation, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for login hook")
	}
}
