package wsbus

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/coder/websocket"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/origins"
)

// Publish enqueues an event for fan-out to subscribers of e.Topic.
// Non-blocking — drops the event if the internal queue is full, which
// only happens under extreme load; the client will refresh on the next
// manual refetch anyway.
func (h *Hub) Publish(e Event) {
	select {
	case h.publish <- e:
	default:
		logger.Global().Warn("wsbus publish queue full, dropping event",
			logger.String("topic", e.Topic),
			logger.String("type", e.Type),
		)
	}
}

func (h *Hub) fanout(ctx context.Context, e Event) {
	msg := serverMessage{Op: "event", Event: &e}
	raw, err := json.Marshal(msg)
	if err != nil {
		logger.Global().Error("wsbus marshal event failed", logger.Error(err))
		return
	}

	h.mu.RLock()
	subscribers := make([]*client, 0, len(h.topics[e.Topic]))
	for c := range h.topics[e.Topic] {
		subscribers = append(subscribers, c)
	}
	h.mu.RUnlock()

	for _, c := range subscribers {
		if !h.authorize(ctx, c.userID, e.Topic) {
			h.removeSubscription(c, e.Topic)
			continue
		}

		// Reacquire the lock before touching c.send. A concurrent unregister
		// may close the channel after the subscriber snapshot above.
		h.mu.RLock()
		_, stillSubscribed := h.topics[e.Topic][c]
		if stillSubscribed {
			select {
			case c.send <- raw:
			default:
				// Client send buffer is full — drop this message for that
				// client rather than block the whole hub. The next event
				// or reconnect will re-sync state.
				logger.Global().Warn("wsbus client send buffer full, dropping",
					logger.Any("user_id", c.userID),
					logger.String("topic", e.Topic),
				)
			}
		}
		h.mu.RUnlock()
	}
}

func (h *Hub) removeSubscription(c *client, topic string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set := h.topics[topic]; set != nil {
		delete(set, c)
		if len(set) == 0 {
			delete(h.topics, topic)
		}
	}
	delete(c.topics, topic)
}

func (h *Hub) shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		c.cancel()
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// HTTP handler
// ─────────────────────────────────────────────────────────────────────────────

// ServeWS upgrades the HTTP request to a websocket and attaches a client.
// The caller must already have authenticated the request and pass the
// resolved userID. We do NOT re-authenticate here; topic subscriptions
// authorize against `userID` via the hub's AuthorizeFn.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request, userID int64) {
	if !h.reserveConnection(userID) {
		logger.Ctx(r.Context()).Warn("wsbus connection rejected: per-user cap",
			logger.Any("user_id", userID), logger.Any("open", maxConnectionsPerUser))
		http.Error(w, "too many connections", http.StatusTooManyRequests)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: origins.AllowedPatternStrings(),
	})
	if err != nil {
		h.releaseConnection(userID)
		logger.Ctx(r.Context()).Warn("wsbus accept failed", logger.Error(err))
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	c := &client{
		userID: userID,
		conn:   conn,
		send:   make(chan []byte, 64),
		topics: make(map[string]struct{}),
		ctx:    ctx,
		cancel: cancel,
	}

	select {
	case h.register <- c:
	case <-r.Context().Done():
		h.releaseConnection(userID)
		cancel()
		_ = conn.Close(websocket.StatusNormalClosure, "")
		return
	}

	// Two goroutines per connection: writer pumps c.send → socket,
	// reader handles inbound subscribe/unsubscribe/ping.
	go h.writeLoop(c)
	h.readLoop(c)

	// readLoop returning means the client disconnected. The unregister
	// handoff is best-effort; on timeout the client's cancelled context
	// marks it for pruneDeadClients so its slot and registration are
	// reclaimed by the hub's periodic sweep.
	select {
	case h.unregister <- c:
	case <-time.After(5 * time.Second):
	}

	_ = conn.Close(websocket.StatusNormalClosure, "")
}

func (h *Hub) readLoop(c *client) {
	defer c.cancel()
	for {
		_, data, err := c.conn.Read(c.ctx)
		if err != nil {
			return
		}
		var msg clientMessage
		if unErr := json.Unmarshal(data, &msg); unErr != nil {
			h.sendErr(c, "invalid json")
			continue
		}

		switch msg.Op {
		case "subscribe":
			if !allowInboundOp(c) {
				h.sendErr(c, "rate limited")
				continue
			}
			topics := msg.Topics
			if len(topics) > maxTopicsPerSubscribeOp {
				topics = topics[:maxTopicsPerSubscribeOp]
			}
			allowed := make([]string, 0, len(topics))
			for _, topic := range topics {
				if h.authorize(c.ctx, c.userID, topic) {
					h.subscribe <- subscription{client: c, topic: topic}
					allowed = append(allowed, topic)
				}
			}
			h.sendOK(c, serverMessage{Op: "subscribed", Topics: allowed})
		case "unsubscribe":
			for _, topic := range msg.Topics {
				h.unsubscribe <- subscription{client: c, topic: topic}
			}
			h.sendOK(c, serverMessage{Op: "unsubscribed", Topics: msg.Topics})
		case "ping":
			h.sendOK(c, serverMessage{Op: "pong"})
		default:
			h.sendErr(c, "unknown op: "+msg.Op)
		}
	}
}

// reserveConnection accounts for a connection before websocket negotiation
// completes. Reserving under the hub lock closes the check-then-increment race
// where concurrent upgrades could all observe the same open count.
func (h *Hub) reserveConnection(userID int64) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.userCap[userID] >= maxConnectionsPerUser {
		return false
	}
	h.userCap[userID]++
	return true
}

func (h *Hub) releaseConnection(userID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.userCap[userID] > 0 {
		h.userCap[userID]--
	}
}

// allowInboundOp rate-limits inbound control frames per connection so a
// single tab cannot drive unbounded authorization queries with oversized
// subscribe batches.
func allowInboundOp(c *client) bool {
	now := time.Now()
	if now.Sub(c.window) >= inboundRateLimitWindow {
		c.window = now
		c.ops = 0
	}
	if c.ops >= maxInboundOpsPerWindow {
		return false
	}
	c.ops++
	return true
}

func (h *Hub) writeLoop(c *client) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			writeCtx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
			err := c.conn.Write(writeCtx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				c.cancel()
				return
			}
		case <-ticker.C:
			// Proactive ping keeps intermediaries (load balancers, nginx)
			// from killing idle connections.
			pingCtx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
			err := c.conn.Ping(pingCtx)
			cancel()
			if err != nil {
				c.cancel()
				return
			}
		}
	}
}

func (h *Hub) sendOK(c *client, msg serverMessage) {
	raw, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case c.send <- raw:
	default:
	}
}

func (h *Hub) sendErr(c *client, reason string) {
	h.sendOK(c, serverMessage{Op: "error", Error: reason})
}
