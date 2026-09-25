package license

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestClockGuardClampsPersistedRollback(t *testing.T) {
	path := filepath.Join(t.TempDir(), "clock.json")
	future := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
	contents := []byte(`{"max_seen":"` + future.Format(time.RFC3339Nano) + `"}`)
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("write clock state: %v", err)
	}

	got := NewClockGuard(path).Now()
	if got.Before(future) {
		t.Fatalf("clock guard returned %s before persisted high-water mark %s", got, future)
	}
}
