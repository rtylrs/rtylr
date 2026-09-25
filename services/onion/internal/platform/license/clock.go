package license

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type clockState struct {
	MaxSeen time.Time `json:"max_seen"`
}

// ClockGuard prevents wall-clock rollback from extending expiry or offline
// grace. The high-water mark survives process restarts in the self-host data
// volume.
type ClockGuard struct {
	path string
	mu   sync.Mutex
	max  time.Time
}

func NewClockGuard(path string) *ClockGuard {
	guard := &ClockGuard{path: path}
	if path == "" {
		return guard
	}
	// #nosec G304 -- the high-water-mark path is deployment configuration and is never derived from a request.
	contents, err := os.ReadFile(path)
	if err != nil {
		return guard
	}
	var state clockState
	if json.Unmarshal(contents, &state) == nil {
		guard.max = state.MaxSeen.UTC()
	}
	return guard
}

func (g *ClockGuard) Now() time.Time {
	now := time.Now().UTC()
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.max.IsZero() && now.Before(g.max) {
		return g.max
	}
	if now.After(g.max) {
		g.max = now
		g.persistLocked()
	}
	return now
}

func (g *ClockGuard) persistLocked() {
	if g.path == "" || g.max.IsZero() {
		return
	}
	if err := os.MkdirAll(filepath.Dir(g.path), 0o700); err != nil {
		return
	}
	contents, err := json.Marshal(clockState{MaxSeen: g.max})
	if err != nil {
		return
	}
	tmp := g.path + ".tmp"
	if err := os.WriteFile(tmp, contents, 0o600); err != nil {
		return
	}
	_ = os.Rename(tmp, g.path)
}
