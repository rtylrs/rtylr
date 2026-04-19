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
func Publish(e Event) {
	h := Default()
	if h == nil {
		return
	}
	h.Publish(e)
}
