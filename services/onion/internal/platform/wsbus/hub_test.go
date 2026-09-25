package wsbus

import (
	"context"
	"testing"
	"time"
)

func TestFanoutRechecksAuthorization(t *testing.T) {
	authorized := true
	calls := 0
	h := New(func(context.Context, int64, string) bool {
		calls++
		return authorized
	})
	c := &client{
		userID: 42,
		send:   make(chan []byte, 4),
		topics: map[string]struct{}{"project:7": {}},
	}
	h.topics["project:7"] = map[*client]struct{}{c: {}}

	// First fan-out authorizes and caches the positive result.
	h.fanout(context.Background(), Event{Topic: "project:7", Type: "task.updated"})
	if got := len(c.send); got != 1 {
		t.Fatalf("authorized fan-out queued %d messages, want 1", got)
	}

	// A revoked membership is not observed until the cached decision expires
	// or is explicitly invalidated; this is what keeps the hot path off the
	// database per subscriber per event (#765).
	authorized = false
	h.fanout(context.Background(), Event{Topic: "project:7", Type: "task.updated"})
	if got := len(c.send); got != 2 {
		t.Fatalf("cached authorization should still deliver; queue length = %d", got)
	}
	if calls != 1 {
		t.Fatalf("authorizer calls = %d, want 1 while cached", calls)
	}

	// Explicit invalidation (or TTL expiry) makes the recheck authoritative.
	h.invalidateAuth(42, "project:7")
	h.fanout(context.Background(), Event{Topic: "project:7", Type: "task.updated"})
	if got := len(c.send); got != 2 {
		t.Fatalf("revoked subscriber received another message; queue length = %d", got)
	}
	if _, ok := h.topics["project:7"][c]; ok {
		t.Fatal("revoked subscriber remained subscribed")
	}
	if calls != 2 {
		t.Fatalf("authorizer calls = %d, want 2 after invalidation", calls)
	}
}

func TestAuthorize_CachesDenialsToo(t *testing.T) {
	calls := 0
	h := New(func(context.Context, int64, string) bool {
		calls++
		return false
	})
	for i := 0; i < 5; i++ {
		if h.authorize(context.Background(), 1, "business:9") {
			t.Fatal("expected deny")
		}
	}
	if calls != 1 {
		t.Fatalf("authorizer calls = %d, want 1 (denial must be cached)", calls)
	}
}

func TestAllowInboundOp_RateLimitWindow(t *testing.T) {
	c := &client{}
	for i := 0; i < maxInboundOpsPerWindow; i++ {
		if !allowInboundOp(c) {
			t.Fatalf("op %d rejected within window", i+1)
		}
	}
	if allowInboundOp(c) {
		t.Fatal("ops beyond the per-window budget must be rejected")
	}
	// A new window resets the budget.
	c.window = time.Now().Add(-inboundRateLimitWindow - time.Second)
	if !allowInboundOp(c) {
		t.Fatal("budget should reset in a fresh window")
	}
}

func TestNilAuthorizerAlwaysAllows(t *testing.T) {
	h := New(nil)
	if !h.authorize(context.Background(), 1, "user:1") {
		t.Fatal("nil authorizer must allow")
	}
}

func TestConnectionCapReservationIsAtomic(t *testing.T) {
	h := New(nil)
	for i := 0; i < maxConnectionsPerUser; i++ {
		if !h.reserveConnection(42) {
			t.Fatalf("reservation %d unexpectedly rejected", i+1)
		}
	}
	if h.reserveConnection(42) {
		t.Fatal("reservation beyond the per-user cap was accepted")
	}
	h.releaseConnection(42)
	if !h.reserveConnection(42) {
		t.Fatal("reservation was not released after a connection closed")
	}
}
