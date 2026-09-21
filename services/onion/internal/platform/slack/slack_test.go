package slack

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/voxire/onion/internal/platform/config"
)

func TestChunkCombinesUnderLimit(t *testing.T) {
	lines := []string{"a", "b", "c"}
	got := chunk(lines, 1000)
	if len(got) != 1 {
		t.Fatalf("expected 1 combined chunk, got %d", len(got))
	}
	for _, l := range lines {
		if !strings.Contains(got[0], l) {
			t.Errorf("chunk missing %q", l)
		}
	}
}

func TestChunkSplitsOverLimit(t *testing.T) {
	long := strings.Repeat("x", 30)
	lines := []string{long, long, long}
	got := chunk(lines, 40) // each line + separator can't all fit in 40
	if len(got) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(got))
	}
	for _, c := range got {
		if len(c) > 40 {
			t.Errorf("chunk exceeds limit: %d chars", len(c))
		}
	}
}

func TestChunkTruncatesOversizedLine(t *testing.T) {
	got := chunk([]string{strings.Repeat("y", 100)}, 40)
	if len(got) != 1 || len(got[0]) > 40 {
		t.Fatalf("oversized line not truncated to limit: %v", len(got[0]))
	}
	if !strings.HasSuffix(got[0], "...") {
		t.Error("expected truncation marker")
	}
}

func TestDecodePostMessageResultSuccess(t *testing.T) {
	got, err := decodePostMessageResult(strings.NewReader(`{"ok":true}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.OK {
		t.Fatal("expected ok result")
	}
}

func TestDecodePostMessageResultLogicalError(t *testing.T) {
	got, err := decodePostMessageResult(strings.NewReader(`{"ok":false,"error":"channel_not_found"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.OK || got.Error != "channel_not_found" {
		t.Fatalf("expected Slack logical error, got %+v", got)
	}
}

func TestDecodePostMessageResultRejectsOversizedBody(t *testing.T) {
	_, err := decodePostMessageResult(strings.NewReader(strings.Repeat("x", maxSlackResponseBodyBytes+1)))
	if !errors.Is(err, errSlackResponseTooLarge) {
		t.Fatalf("expected errSlackResponseTooLarge, got %v", err)
	}
}

// enableSlackForTest arms the package for a lifecycle test and restores the
// previous config when the test ends.
func enableSlackForTest(t *testing.T) {
	t.Helper()
	prev := config.GlobalConfig
	config.GlobalConfig = &config.AppConfig{
		SlackLoggingEnabled: true,
		SlackBotToken:       "xoxb-test",
		SlackAlertsChannel:  "C_ALERTS",
	}
	t.Cleanup(func() {
		Shutdown()
		config.GlobalConfig = prev
	})
}

func currentNotifier(t *testing.T) *notifier {
	t.Helper()
	mu.RLock()
	defer mu.RUnlock()
	if global == nil {
		t.Fatal("expected an initialized notifier")
	}
	return global
}

// After Shutdown the stopped notifier must no longer be reachable: otherwise
// Enabled() stays true and Post keeps buffering into a loop that has exited.
func TestShutdownClearsStoppedNotifier(t *testing.T) {
	enableSlackForTest(t)
	Init()

	n := currentNotifier(t)
	if !Enabled() {
		t.Fatal("Enabled() = false right after Init")
	}

	Shutdown()

	if Enabled() {
		t.Error("Enabled() = true after Shutdown; a stopped notifier is still reachable")
	}
	Post(Alerts, "queued after shutdown")

	n.mu.Lock()
	buffered := len(n.bufs[Alerts])
	n.mu.Unlock()
	if buffered != 0 {
		t.Errorf("Post buffered %d message(s) into a stopped notifier, where they would never flush", buffered)
	}
}

// A later Init has to arm a fresh, working notifier rather than leave the
// stopped one in place.
func TestInitAfterShutdownRearms(t *testing.T) {
	enableSlackForTest(t)
	Init()
	Shutdown()

	Init()

	n := currentNotifier(t)
	if !Enabled() {
		t.Fatal("Enabled() = false after re-Init")
	}
	if n.stopped {
		t.Fatal("re-Init reused the stopped notifier")
	}

	rec := &recordingTransport{}
	n.client = &http.Client{Transport: rec}
	Post(Alerts, "alert after restart")
	n.flush()
	if got := rec.calls(); got != 1 {
		t.Errorf("flush after re-Init posted %d message(s), want 1", got)
	}
}

type recordingTransport struct {
	mu    sync.Mutex
	count int
}

func (r *recordingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	r.mu.Lock()
	r.count++
	r.mu.Unlock()
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
		Header:     make(http.Header),
	}, nil
}

func (r *recordingTransport) calls() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.count
}
