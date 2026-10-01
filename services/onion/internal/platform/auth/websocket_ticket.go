package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const (
	webSocketTicketLifetime  = 30 * time.Second
	webSocketTicketKeyPrefix = "auth:websocket-ticket:"
)

var (
	ErrWebSocketTicketInvalid = errors.New("websocket ticket invalid")
	ErrWebSocketTicketStorage = errors.New("websocket ticket storage failed")
)

type websocketTicketRecord struct {
	UserID         int64  `json:"user_id"`
	BusinessID     int64  `json:"business_id"`
	IsConsoleAdmin bool   `json:"is_console_admin,omitempty"`
	Platform       string `json:"platform,omitempty"`
	PlatformUserID int64  `json:"platform_user_id,omitempty"`
	PlanID         int64  `json:"plan_id,omitempty"`
	ExpiresAt      int64  `json:"expires_at_unix_nano"`
}

func (r websocketTicketRecord) authUser() *AuthUser {
	return &AuthUser{
		UserID:         r.UserID,
		BusinessID:     r.BusinessID,
		IsConsoleAdmin: r.IsConsoleAdmin,
		Platform:       r.Platform,
		PlatformUserID: r.PlatformUserID,
		PlanID:         r.PlanID,
	}
}

// The local map is the development/test fallback when Redis is disabled. Its
// keys are hashes, never the bearer or ticket values themselves. Production
// deployments with Redis use Redis GETDEL for cross-instance atomic consume.
var websocketTickets = struct {
	sync.Mutex
	entries map[string]websocketTicketRecord
}{
	entries: make(map[string]websocketTicketRecord),
}

// IssueWebSocketTicket creates an opaque, short-lived ticket for an already
// authenticated user. Only the ticket hash and claims needed by the realtime
// handler are stored; the user's JWT is never copied into the ticket record.
func IssueWebSocketTicket(ctx context.Context, user *AuthUser) (string, error) {
	if user == nil || user.UserID == 0 {
		return "", ErrWebSocketUserMissing
	}
	return issueWebSocketTicketAt(ctx, user, time.Now())
}

func issueWebSocketTicketAt(ctx context.Context, user *AuthUser, issuedAt time.Time) (string, error) {
	if user == nil || user.UserID == 0 {
		return "", ErrWebSocketUserMissing
	}

	rawTicket := make([]byte, 32)
	if _, err := rand.Read(rawTicket); err != nil {
		return "", err
	}

	ticket := base64.RawURLEncoding.EncodeToString(rawTicket)
	record := websocketTicketRecord{
		UserID:         user.UserID,
		BusinessID:     user.BusinessID,
		IsConsoleAdmin: user.IsConsoleAdmin,
		Platform:       user.Platform,
		PlatformUserID: user.PlatformUserID,
		PlanID:         user.PlanID,
		ExpiresAt:      issuedAt.Add(webSocketTicketLifetime).UnixNano(),
	}
	if err := storeWebSocketTicket(ctx, ticket, record); err != nil {
		return "", err
	}

	return ticket, nil
}

// storeWebSocketTicket stores a ticket for later consumption. Redis is the
// production source of truth with cross-instance atomic GETDEL; the
// process-local map is only a fallback for deployments and tests without Redis.
// When Redis is selected, a failed write is returned to the caller rather than
// issuing a ticket that no handshake can consume.
func storeWebSocketTicket(ctx context.Context, ticket string, record websocketTicketRecord) error {
	return storeWebSocketTicketWithPut(
		ctx,
		ticket,
		record,
		cacher.RedisAvailable(),
		func(ctx context.Context, key string, value []byte, ttl time.Duration) bool {
			cache := cacher.GetRedisCache()
			return cache != nil && cache.PutConfirmed(ctx, key, value, ttl)
		},
	)
}

// storeWebSocketTicketWithPut keeps the storage decision injectable for a
// regression test without requiring a live Redis server.
func storeWebSocketTicketWithPut(
	ctx context.Context,
	ticket string,
	record websocketTicketRecord,
	useRedis bool,
	put func(context.Context, string, []byte, time.Duration) bool,
) error {
	hash := hashWebSocketTicket(ticket)

	if !useRedis {
		storeLocalWebSocketTicket(hash, record)
		return nil
	}

	encoded, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal websocket ticket: %w", err)
	}

	remaining := time.Until(time.Unix(0, record.ExpiresAt))
	if remaining <= 0 {
		return ErrWebSocketTicketStorage
	}
	if put == nil || !put(ctx, websocketTicketKey(hash), encoded, remaining) {
		return ErrWebSocketTicketStorage
	}

	return nil
}

// storeLocalWebSocketTicket records a ticket in the fallback map and drops
// already-expired entries, so an abandoned ticket cannot accumulate in the map
// for the lifetime of the process.
func storeLocalWebSocketTicket(hash string, record websocketTicketRecord) {
	now := time.Now().UnixNano()
	websocketTickets.Lock()
	for key, existing := range websocketTickets.entries {
		if existing.ExpiresAt <= now {
			delete(websocketTickets.entries, key)
		}
	}
	websocketTickets.entries[hash] = record
	websocketTickets.Unlock()
}

// ConsumeWebSocketTicket atomically consumes a ticket and returns its user.
// Redis GETDEL is used when Redis is enabled so a ticket cannot be replayed
// across API instances. The mutex-protected map provides the same one-use
// semantics for single-process development and tests.
func ConsumeWebSocketTicket(ctx context.Context, ticket string) (*AuthUser, bool) {
	ticket = strings.TrimSpace(ticket)
	if ticket == "" {
		return nil, false
	}

	hash := hashWebSocketTicket(ticket)
	if cacher.RedisAvailable() {
		encoded := cacher.GetRedisCache().GetDel(ctx, websocketTicketKey(hash))
		deleteLocalWebSocketTicket(hash)
		if len(encoded) == 0 {
			return nil, false
		}

		var record websocketTicketRecord
		if err := json.Unmarshal(encoded, &record); err != nil || record.ExpiresAt <= time.Now().UnixNano() {
			return nil, false
		}
		if record.UserID == 0 {
			return nil, false
		}

		return record.authUser(), true
	}

	websocketTickets.Lock()
	record, ok := websocketTickets.entries[hash]
	delete(websocketTickets.entries, hash)
	websocketTickets.Unlock()
	if !ok || record.ExpiresAt <= time.Now().UnixNano() || record.UserID == 0 {
		return nil, false
	}

	return record.authUser(), true
}

func deleteLocalWebSocketTicket(hash string) {
	websocketTickets.Lock()
	delete(websocketTickets.entries, hash)
	websocketTickets.Unlock()
}

func hashWebSocketTicket(ticket string) string {
	digest := sha256.Sum256([]byte(ticket))
	return hex.EncodeToString(digest[:])
}

func websocketTicketKey(hash string) string {
	return webSocketTicketKeyPrefix + hash
}
