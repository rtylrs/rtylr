package wsbus

import (
	"context"
	"testing"
)

func TestFanoutRechecksAuthorization(t *testing.T) {
	authorized := true
	h := New(func(context.Context, int64, string) bool { return authorized })
	c := &client{
		userID: 42,
		send:   make(chan []byte, 1),
		topics: map[string]struct{}{"project:7": {}},
	}
	h.topics["project:7"] = map[*client]struct{}{c: {}}

	h.fanout(context.Background(), Event{Topic: "project:7", Type: "task.updated"})
	if got := len(c.send); got != 1 {
		t.Fatalf("authorized fan-out queued %d messages, want 1", got)
	}

	authorized = false
	h.fanout(context.Background(), Event{Topic: "project:7", Type: "task.updated"})
	if got := len(c.send); got != 1 {
		t.Fatalf("revoked subscriber received another message; queue length = %d", got)
	}
	if _, ok := h.topics["project:7"][c]; ok {
		t.Fatal("revoked subscriber remained subscribed")
	}
}
