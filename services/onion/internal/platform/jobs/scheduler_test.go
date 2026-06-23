package jobs

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type fakeLocker struct {
	acquire int32
	release int32
	ok      bool
	err     error
}

func (l *fakeLocker) TryLock(context.Context, string) (func(context.Context) error, bool, error) {
	atomic.AddInt32(&l.acquire, 1)
	if l.err != nil || !l.ok {
		return nil, l.ok, l.err
	}
	return func(context.Context) error {
		atomic.AddInt32(&l.release, 1)
		return nil
	}, true, nil
}

func TestSchedulerStartRejectsInvalidSchedule(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register(Job{
		Name:     "bad",
		Schedule: "not a cron",
		Handler:  func(context.Context) error { return nil },
	}); err != nil {
		t.Fatal(err)
	}

	err := NewScheduler(registry).Start(context.Background())
	if err == nil {
		t.Fatal("expected invalid schedule error")
	}
}

func TestSchedulerRunSkipsDuplicateLocalRun(t *testing.T) {
	registry := NewRegistry()
	started := make(chan struct{})
	release := make(chan struct{})
	if err := registry.Register(Job{
		Name: "slow",
		Handler: func(context.Context) error {
			close(started)
			<-release
			return nil
		},
	}); err != nil {
		t.Fatal(err)
	}

	locker := &fakeLocker{ok: true}
	scheduler := NewScheduler(registry, WithLocker(locker), WithDefaultTimeout(time.Minute))
	done := make(chan error, 1)
	go func() {
		_, err := scheduler.Run(context.Background(), "slow")
		done <- err
	}()

	<-started
	_, err := scheduler.Run(context.Background(), "slow")
	if !errors.Is(err, ErrJobAlreadyRunning) {
		t.Fatalf("Run duplicate err = %v, want ErrJobAlreadyRunning", err)
	}

	close(release)
	if err := <-done; err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if got := atomic.LoadInt32(&locker.acquire); got != 1 {
		t.Fatalf("lock acquire count = %d, want 1", got)
	}
	if got := atomic.LoadInt32(&locker.release); got != 1 {
		t.Fatalf("lock release count = %d, want 1", got)
	}
}

func TestSchedulerRunSkipsWhenDistributedLockUnavailable(t *testing.T) {
	registry := NewRegistry()
	var ran atomic.Bool
	if err := registry.Register(Job{
		Name: "locked",
		Handler: func(context.Context) error {
			ran.Store(true)
			return nil
		},
	}); err != nil {
		t.Fatal(err)
	}

	_, err := NewScheduler(registry, WithLocker(&fakeLocker{})).Run(context.Background(), "locked")
	if !errors.Is(err, ErrJobLockNotAcquired) {
		t.Fatalf("Run err = %v, want ErrJobLockNotAcquired", err)
	}
	if ran.Load() {
		t.Fatal("handler ran even though distributed lock was unavailable")
	}
}

func TestSchedulerEnvironmentHelpers(t *testing.T) {
	t.Setenv(EnvSchedulerEnabled, "false")
	if SchedulerEnabledFromEnv() {
		t.Fatal("scheduler should be disabled when SCHEDULER_ENABLED=false")
	}

	t.Setenv(EnvSchedulerEnabled, "true")
	t.Setenv(EnvSchedulerDisabled, "true")
	if SchedulerEnabledFromEnv() {
		t.Fatal("scheduler disabled flag should win")
	}

	t.Setenv("TEST_JOB_DURATION", "45s")
	if got := DurationFromEnv("TEST_JOB_DURATION", time.Second); got != 45*time.Second {
		t.Fatalf("duration = %s, want 45s", got)
	}
}
