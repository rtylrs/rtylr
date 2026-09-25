package wsbus

import (
	"context"
	"sync"
)

// defaultHub is the process-wide hub. Publishers across modules use
// Default() to emit invalidation events without plumbing a hub instance
// through every service constructor.
var (
	defaultHub *Hub
	defaultMu  sync.Mutex
)

// SetDefault replaces the process-wide hub and starts its run loop on
// the provided context. Call once at server startup from main.go:
//
//	wsbus.SetDefault(ctx, wsbus.New(myAuthFn))
func SetDefault(ctx context.Context, h *Hub) {
	defaultMu.Lock()
	defaultHub = h
	defaultMu.Unlock()
	go h.Run(ctx)
}

// Default returns the process-wide hub or nil if one has not been set.
// Publishers MUST tolerate a nil hub (Publish on nil is a no-op) so
// production reads are never blocked by optional realtime plumbing.
func Default() *Hub {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	return defaultHub
}

// Publish is the ergonomic helper for the common publish-via-default
// pattern. Silently drops events if no hub is configured.
//
// When the Redis backplane is enabled (see EnableRedisBackplane) the event is
// published to Redis instead of fanned out locally; every replica's subscriber
// — including this one — then delivers it to its own hub, so clients connected
// to any replica receive it exactly once. Without the backplane it fans out
// in-process only (correct for a single replica).
func Publish(e Event) {
	if backplaneEnabled.Load() {
		publishBackplane(e)
		return
	}
	h := Default()
	if h == nil {
		return
	}
	h.Publish(e)
}

// ParseTopic splits a "<scope>:<id>" topic into its parts. Returns
// (scope, id, true) when the shape matches and id parses as int64;
// (_, _, false) otherwise. Callers use this to dispatch authorization
// rules by scope without repeating the parse logic.
func ParseTopic(topic string) (string, int64, bool) {
	colon := -1
	for i := 0; i < len(topic); i++ {
		if topic[i] == ':' {
			colon = i
			break
		}
	}
	if colon <= 0 || colon == len(topic)-1 {
		return "", 0, false
	}
	var id int64
	for i := colon + 1; i < len(topic); i++ {
		c := topic[i]
		if c < '0' || c > '9' {
			return "", 0, false
		}
		id = id*10 + int64(c-'0')
	}
	return topic[:colon], id, true
}
