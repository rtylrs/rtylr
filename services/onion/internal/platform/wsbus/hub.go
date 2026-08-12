// Package wsbus is a generic websocket pub/sub hub used for realtime fan-out
// across Flow (and any other module that needs it).
//
// The design on purpose does NOT send full entity payloads. Events are
// lightweight invalidation hints — `{topic, type, entity_id}` — and the
// client refetches canonical data through the existing REST endpoints.
// This keeps one source of truth for the wire format (the handler + proto)
// and avoids drift between the WS payload and the HTTP payload.
//
// Topics are plain strings formatted as "<scope>:<id>":
//
//	user:42          — personal events (notifications, invites)
//	workspace:12     — workspace-level events (member added, etc.)
//	project:7        — task changes in that project (created/updated/moved)
//	task:88          — comments + activity on a specific task
//
// Authorization is enforced on subscribe and again before each event is
// fanned out. Rechecking at fan-out prevents a client whose membership was
// revoked after subscribing from receiving later invalidation events.
package wsbus

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/origins"
)

// AuthorizeFn decides whether userID may subscribe to topic. Return true
// to allow. Implementations typically check workspace/project membership.
type AuthorizeFn func(ctx context.Context, userID int64, topic string) bool

// Event is the payload broadcast to all clients subscribed to Topic.
// Type is a dotted string like "task.updated" or "notification.created".
// EntityID carries the id of the affected record so the client can
// invalidate a precise react-query key instead of the whole list.
type Event struct {
	Topic    string `json:"topic"`
	Type     string `json:"type"`
	EntityID int64  `json:"entity_id,omitempty"`
	// ActorID is the user who caused the event; the client uses it to
	// skip refetches triggered by its own writes (the write already
	// invalidated locally), avoiding a redundant round-trip.
	ActorID int64 `json:"actor_id,omitempty"`
}

// clientMessage is what the browser sends to us.
type clientMessage struct {
	Op     string   `json:"op"`
	Topics []string `json:"topics,omitempty"`
}

// serverMessage is what we send back.
type serverMessage struct {
	Op     string   `json:"op"`
	Topics []string `json:"topics,omitempty"`
	Event  *Event   `json:"event,omitempty"`
	Error  string   `json:"error,omitempty"`
}

// Hub is a running websocket broker. Construct with New, start the
// background loop with Run, then attach the HTTP handler.
type Hub struct {
	auth AuthorizeFn

	mu      sync.RWMutex
	topics  map[string]map[*client]struct{} // topic -> set of clients
	clients map[*client]struct{}

	publish     chan Event
	subscribe   chan subscription
	unsubscribe chan subscription
	register    chan *client
	unregister  chan *client
}

type subscription struct {
	client *client
	topic  string
}

type client struct {
	userID int64
	conn   *websocket.Conn
	send   chan []byte
	topics map[string]struct{}

	ctx    context.Context
	cancel context.CancelFunc
}

// New returns a hub with the given per-topic authorizer.
func New(auth AuthorizeFn) *Hub {
	return &Hub{
		auth:        auth,
		topics:      make(map[string]map[*client]struct{}),
		clients:     make(map[*client]struct{}),
		publish:     make(chan Event, 256),
		subscribe:   make(chan subscription, 64),
		unsubscribe: make(chan subscription, 64),
		register:    make(chan *client, 16),
		unregister:  make(chan *client, 16),
	}
}

// Run drives the hub's internal loop. Block until ctx is cancelled.
// Call in a goroutine at server startup.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			h.shutdown()
			return

		case c := <-h.register:
			h.mu.Lock()
			h.clients[c] = struct{}{}
			h.mu.Unlock()

		case c := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				for topic := range c.topics {
					if set := h.topics[topic]; set != nil {
						delete(set, c)
						if len(set) == 0 {
							delete(h.topics, topic)
						}
					}
				}
				close(c.send)
			}
			h.mu.Unlock()

		case s := <-h.subscribe:
			h.mu.Lock()
			if h.topics[s.topic] == nil {
				h.topics[s.topic] = make(map[*client]struct{})
			}
			h.topics[s.topic][s.client] = struct{}{}
			s.client.topics[s.topic] = struct{}{}
			h.mu.Unlock()

		case s := <-h.unsubscribe:
			h.mu.Lock()
			if set := h.topics[s.topic]; set != nil {
				delete(set, s.client)
				if len(set) == 0 {
					delete(h.topics, s.topic)
				}
			}
			delete(s.client.topics, s.topic)
			h.mu.Unlock()

		case e := <-h.publish:
			h.fanout(ctx, e)
		}
	}
}

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
		if h.auth != nil && !h.auth(ctx, c.userID, e.Topic) {
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
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: origins.AllowedPatternStrings(),
	})
	if err != nil {
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

	h.register <- c

	// Two goroutines per connection: writer pumps c.send → socket,
	// reader handles inbound subscribe/unsubscribe/ping.
	go h.writeLoop(c)
	h.readLoop(c)

	// readLoop returning means the client disconnected.
	h.unregister <- c
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
			allowed := make([]string, 0, len(msg.Topics))
			for _, topic := range msg.Topics {
				if h.auth == nil || h.auth(c.ctx, c.userID, topic) {
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
