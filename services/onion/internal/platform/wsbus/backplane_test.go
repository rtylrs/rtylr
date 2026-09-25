package wsbus

import (
	"context"
	"sync"
	"testing"
	"time"
)

func resetBackplaneForTest() {
	backplaneMu.Lock()
	if backplaneCancel != nil {
		backplaneCancel()
	}
	backplaneCancel = nil
	backplaneGeneration = 0
	backplaneMu.Unlock()
	backplaneEnabled.Store(false)
}

func waitForBackplaneDisabled(t *testing.T) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if !backplaneEnabled.Load() {
			return
		}
		time.Sleep(time.Millisecond)
	}

	t.Fatal("backplane stayed enabled after its subscriber context was canceled")
}

// Regression: every in-process server restart re-ran setup routes and started
// another Redis subscriber while the previous one (bound to
// context.Background()) kept running, so each later event was fanned into the
// live hub once per leaked subscriber.
func TestEnableRedisBackplane_ReplacesPreviousSubscriber(t *testing.T) {
	resetBackplaneForTest()
	t.Cleanup(resetBackplaneForTest)

	var mu sync.Mutex
	var subscriberCtxs []context.Context

	original := startBackplaneSubscriber
	startBackplaneSubscriber = func(ctx context.Context) {
		mu.Lock()
		defer mu.Unlock()

		subscriberCtxs = append(subscriberCtxs, ctx)
	}
	t.Cleanup(func() { startBackplaneSubscriber = original })

	firstRun, cancelFirst := context.WithCancel(context.Background())
	defer cancelFirst()
	EnableRedisBackplane(firstRun)

	if !backplaneEnabled.Load() {
		t.Fatal("backplane should be enabled after EnableRedisBackplane")
	}

	secondRun, cancelSecond := context.WithCancel(context.Background())
	EnableRedisBackplane(secondRun)

	mu.Lock()
	got := append([]context.Context(nil), subscriberCtxs...)
	mu.Unlock()

	if len(got) != 2 {
		t.Fatalf("expected 2 subscribers (one per call), got %d", len(got))
	}
	if got[0].Err() == nil {
		t.Error("previous subscriber's context is still live; subscribers stack on every restart")
	}
	if got[1].Err() != nil {
		t.Error("new subscriber's context was already canceled")
	}

	// When a server run ends without a replacement, publishes must fall back to
	// in-process fan-out instead of going into a channel nobody reads here.
	cancelSecond()
	waitForBackplaneDisabled(t)
}
