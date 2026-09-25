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
	"strconv"
	"sync"
	"time"

	"github.com/coder/websocket"
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
// Authorization results are cached per (user, topic) so the hub's single
// goroutine does not run a membership DB query per subscriber per broadcast
// event (#765). Revocation therefore takes effect within defaultAuthTTL.
const (
	defaultAuthTTL          = 45 * time.Second
	maxTopicsPerSubscribeOp = 64
	maxTopicsPerClient      = 256
	maxConnectionsPerUser   = 8
	maxInboundOpsPerWindow  = 100
	inboundRateLimitWindow  = time.Second
)

type authCacheEntry struct {
	allowed bool
	expires time.Time
}

type Hub struct {
	auth    AuthorizeFn
	authTTL time.Duration

	authMu    sync.RWMutex
	authCache map[string]authCacheEntry

	mu      sync.RWMutex
	topics  map[string]map[*client]struct{} // topic -> set of clients
	clients map[*client]struct{}
	userCap map[int64]int

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
	ops    int
	window time.Time

	ctx    context.Context
	cancel context.CancelFunc
}

// New returns a hub with the given per-topic authorizer.
func New(auth AuthorizeFn) *Hub {
	return &Hub{
		auth:        auth,
		authTTL:     defaultAuthTTL,
		authCache:   make(map[string]authCacheEntry),
		topics:      make(map[string]map[*client]struct{}),
		clients:     make(map[*client]struct{}),
		userCap:     make(map[int64]int),
		publish:     make(chan Event, 256),
		subscribe:   make(chan subscription, 64),
		unsubscribe: make(chan subscription, 64),
		register:    make(chan *client, 16),
		unregister:  make(chan *client, 16),
	}
}

// authorize consults the per-(user, topic) result cache before invoking the
// production authorizer, which typically costs a MySQL membership query.
// Both allow and deny results are cached for the TTL.
func (h *Hub) authorize(ctx context.Context, userID int64, topic string) bool {
	if h.auth == nil {
		return true
	}

	key := strconv.FormatInt(userID, 10) + "|" + topic
	now := time.Now()

	h.authMu.RLock()
	entry, ok := h.authCache[key]
	h.authMu.RUnlock()
	if ok && now.Before(entry.expires) {
		return entry.allowed
	}

	allowed := h.auth(ctx, userID, topic)

	h.authMu.Lock()
	h.authCache[key] = authCacheEntry{allowed: allowed, expires: now.Add(h.authTTL)}
	h.authMu.Unlock()

	return allowed
}

// invalidateAuth drops any cached decision for (userID, topic) so the next
// check re-consults the authorizer immediately.
func (h *Hub) invalidateAuth(userID int64, topic string) {
	key := strconv.FormatInt(userID, 10) + "|" + topic
	h.authMu.Lock()
	delete(h.authCache, key)
	h.authMu.Unlock()
}

func (h *Hub) pruneAuthCache() {
	now := time.Now()
	h.authMu.Lock()
	for key, entry := range h.authCache {
		if now.After(entry.expires) {
			delete(h.authCache, key)
		}
	}
	h.authMu.Unlock()
}

// pruneDeadClients reaps clients whose context is already cancelled but
// whose unregister handoff never reached the hub (e.g. ServeWS timed out
// waiting on the unregister channel). Without this, the per-user connection
// slot and client registration would leak until restart.
func (h *Hub) pruneDeadClients() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		if c.ctx.Err() == nil {
			continue
		}
		delete(h.clients, c)
		if h.userCap[c.userID] > 0 {
			h.userCap[c.userID]--
		}
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
}

// Run drives the hub's internal loop. Block until ctx is cancelled.
// Call in a goroutine at server startup.
func (h *Hub) Run(ctx context.Context) {
	prune := time.NewTicker(time.Minute)
	defer prune.Stop()

	for {
		select {
		case <-ctx.Done():
			h.shutdown()
			return

		case <-prune.C:
			h.pruneAuthCache()
			h.pruneDeadClients()

		case c := <-h.register:
			h.mu.Lock()
			h.clients[c] = struct{}{}
			h.mu.Unlock()

		case c := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				if h.userCap[c.userID] > 0 {
					h.userCap[c.userID]--
				}
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
			if _, connected := h.clients[s.client]; !connected {
				h.mu.Unlock()
				continue
			}
			if _, alreadySubscribed := s.client.topics[s.topic]; alreadySubscribed || len(s.client.topics) >= maxTopicsPerClient {
				h.mu.Unlock()
				continue
			}
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
