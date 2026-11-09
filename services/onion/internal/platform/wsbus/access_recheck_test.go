package wsbus

import (
	"context"
	"testing"
)

// A cached "allowed" kept delivering a project's events to someone removed
// from it until the cache expired (up to defaultAuthTTL). An access recheck
// event re-runs the authorizer for the affected subscriptions at once.

func subscribed(h *Hub, userID int64, topics ...string) *client {
	c := &client{userID: userID, send: make(chan []byte, 8), topics: map[string]struct{}{}}
	for _, topic := range topics {
		c.topics[topic] = struct{}{}
		if h.topics[topic] == nil {
			h.topics[topic] = map[*client]struct{}{}
		}
		h.topics[topic][c] = struct{}{}
	}
	h.clients[c] = struct{}{}
	return c
}

func TestAccessRecheckForAUserEndsWhatTheyMayNoLongerSee(t *testing.T) {
	allowed := map[string]bool{"project:7": true, "task:70": true, "user:42": true}
	h := New(func(_ context.Context, _ int64, topic string) bool { return allowed[topic] })
	removed := subscribed(h, 42, "project:7", "task:70", "user:42")
	other := subscribed(h, 43, "project:7")

	// Warm the cache the way normal traffic does.
	h.fanout(context.Background(), Event{Topic: "project:7", Type: "task.updated"})

	allowed["project:7"], allowed["task:70"] = false, false
	h.handle(context.Background(), AccessRecheckEvent("user:42"))

	if _, ok := removed.topics["project:7"]; ok {
		t.Error("still subscribed to the project after losing access")
	}
	if _, ok := removed.topics["task:70"]; ok {
		t.Error("still subscribed to the project's task after losing access")
	}
	if _, ok := removed.topics["user:42"]; !ok {
		t.Error("lost a subscription that is still allowed")
	}
	if _, ok := h.topics["project:7"][other]; !ok {
		t.Error("another user's subscription was touched by a recheck for user 42")
	}
	// The control event itself is never delivered to anyone.
	if len(removed.send) != 1 || len(other.send) != 1 {
		t.Errorf("queued messages = %d and %d, want only the one task event each", len(removed.send), len(other.send))
	}
}

func TestAccessRecheckForATopicCoversEveryoneWatchingIt(t *testing.T) {
	member := map[int64]bool{42: true, 43: true}
	h := New(func(_ context.Context, userID int64, _ string) bool { return member[userID] })
	stays := subscribed(h, 42, "project:7")
	leaves := subscribed(h, 43, "project:7", "task:70")
	h.fanout(context.Background(), Event{Topic: "project:7", Type: "task.updated"})

	member[43] = false
	h.handle(context.Background(), AccessRecheckEvent("project:7"))

	if _, ok := h.topics["project:7"][stays]; !ok {
		t.Error("a member lost their subscription")
	}
	if len(leaves.topics) != 0 {
		t.Errorf("a user who lost access is still subscribed to %v", leaves.topics)
	}
}
