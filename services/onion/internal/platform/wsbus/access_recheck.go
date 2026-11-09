package wsbus

import "context"

// accessRecheckType marks a control event. It is never delivered: a hub that
// receives one re-runs the authorizer at once for the subscriptions it names
// and ends those it no longer allows. Access decisions are otherwise cached
// for defaultAuthTTL, so without it someone removed from a project kept
// getting its events for up to that long.
const accessRecheckType = "wsbus.access_recheck"

// AccessRecheckEvent builds the control event for scope. "user:<id>" rechecks
// every subscription that user holds; any other topic rechecks every
// subscription of each client watching it (their task topics included, which
// a project's access also governs).
func AccessRecheckEvent(scope string) Event {
	return Event{Topic: scope, Type: accessRecheckType}
}

// RecheckAccess asks every hub, on every instance through the backplane, to
// recheck scope now. Call it after a change that can take access away.
func RecheckAccess(scope string) {
	Publish(AccessRecheckEvent(scope))
}

// handle is what the hub loop does with each published event.
func (h *Hub) handle(ctx context.Context, e Event) {
	if e.Type == accessRecheckType {
		h.recheck(ctx, e.Topic)
		return
	}
	h.fanout(ctx, e)
}

func (h *Hub) recheck(ctx context.Context, scope string) {
	type held struct {
		c     *client
		topic string
	}
	kind, userID, isUser := ParseTopic(scope)
	isUser = isUser && kind == "user"

	h.mu.RLock()
	var targets []*client
	if isUser {
		for c := range h.clients {
			if c.userID == userID {
				targets = append(targets, c)
			}
		}
	} else {
		for c := range h.topics[scope] {
			targets = append(targets, c)
		}
	}
	var subscriptions []held
	for _, c := range targets {
		for topic := range c.topics {
			subscriptions = append(subscriptions, held{c, topic})
		}
	}
	h.mu.RUnlock()

	for _, s := range subscriptions {
		h.invalidateAuth(s.c.userID, s.topic)
		if !h.authorize(ctx, s.c.userID, s.topic) {
			h.removeSubscription(s.c, s.topic)
		}
	}
}
