package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
)

func signedWebSocketToken(t *testing.T, claims jwtutils.Claims) string {
	t.Helper()

	token, err := jwtutils.NewJWTHelperWithDuration("ws-secret", webSocketJWTIssuer, time.Hour).CreateJWT(claims)
	if err != nil {
		t.Fatalf("CreateJWT: %v", err)
	}
	return token
}

func wsRequest(path string) *http.Request {
	ctx := context.WithValue(context.Background(), jwtutils.JWTSecretKey, "ws-secret")
	return httptest.NewRequestWithContext(ctx, http.MethodGet, path, http.NoBody)
}

// forceLocalWebSocketTicketStore makes the fallback contract explicit for unit
// tests that inspect the process-local store. Redis is selected during package
// initialization, before a test can set its environment, so REDIS_ENABLED=true
// must not silently change this test's storage contract.
func forceLocalWebSocketTicketStore(t *testing.T) {
	t.Helper()
	if cacher.RedisAvailable() {
		cacher.DisableRedisForProcess(context.Background())
	}

	websocketTickets.Lock()
	for hash := range websocketTickets.entries {
		delete(websocketTickets.entries, hash)
	}
	websocketTickets.Unlock()
}

func TestAuthUserFromWebSocketRequest_Ticket(t *testing.T) {
	forceLocalWebSocketTicketStore(t)
	ticket, err := IssueWebSocketTicket(context.Background(), &AuthUser{
		UserID:         7,
		BusinessID:     3,
		Platform:       "rtylr",
		PlatformUserID: 42,
		PlanID:         9,
	})
	if err != nil {
		t.Fatalf("IssueWebSocketTicket: %v", err)
	}

	req := wsRequest("/flow/ws?ticket=" + url.QueryEscape(ticket))
	req.RemoteAddr = "1.2.3.4:12345"

	user, err := AuthUserFromWebSocketRequest(req)
	if err != nil {
		t.Fatalf("AuthUserFromWebSocketRequest: %v", err)
	}
	if user.UserID != 7 || user.BusinessID != 3 {
		t.Fatalf("unexpected auth user: %+v", user)
	}
	if user.Platform != "rtylr" || user.PlatformUserID != 42 || user.PlanID != 9 {
		t.Fatalf("platform claims were not preserved: %+v", user)
	}
	if user.Token != "" || user.IP != "1.2.3.4:12345" {
		t.Fatalf("ticket auth should not retain a bearer token: %+v", user)
	}
}

func TestAuthUserFromWebSocketRequest_QueryTokenRejected(t *testing.T) {
	token := signedWebSocketToken(t, &webSocketClaims{UserID: 7})
	req := wsRequest("/flow/ws?token=" + token)

	_, err := AuthUserFromWebSocketRequest(req)
	if !errors.Is(err, ErrWebSocketTokenMissing) {
		t.Fatalf("expected legacy query token to be rejected, got %v", err)
	}
}

func TestAuthUserFromWebSocketRequest_NonFlowQueryTokenRejected(t *testing.T) {
	for _, path := range []string{"/dash/ws", "/pos/ws"} {
		t.Run(path, func(t *testing.T) {
			token := signedWebSocketToken(t, &webSocketClaims{UserID: 17, BusinessID: 4})
			req := wsRequest(path + "?token=" + token)

			_, err := AuthUserFromWebSocketRequest(req)
			if !errors.Is(err, ErrWebSocketTokenMissing) {
				t.Fatalf("expected legacy query token to be rejected for %s, got %v", path, err)
			}
		})
	}
}

func TestAuthUserFromWebSocketRequest_BearerFallback(t *testing.T) {
	token := signedWebSocketToken(t, &webSocketClaims{UserID: 8, BusinessID: 4})
	req := wsRequest("/dash/ws")
	req.Header.Set("Authorization", "Bearer "+token)

	user, err := AuthUserFromWebSocketRequest(req)
	if err != nil {
		t.Fatalf("AuthUserFromWebSocketRequest: %v", err)
	}
	if user.UserID != 8 || user.BusinessID != 4 {
		t.Fatalf("unexpected auth user: %+v", user)
	}
}

func TestAuthUserFromWebSocketRequest_MissingToken(t *testing.T) {
	_, err := AuthUserFromWebSocketRequest(wsRequest("/pos/ws"))
	if !errors.Is(err, ErrWebSocketTokenMissing) {
		t.Fatalf("expected ErrWebSocketTokenMissing, got %v", err)
	}
}

func TestAuthUserFromWebSocketRequest_MissingSecret(t *testing.T) {
	token := signedWebSocketToken(t, &webSocketClaims{UserID: 1})
	req := httptest.NewRequest(http.MethodGet, "/pos/ws", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)

	_, err := AuthUserFromWebSocketRequest(req)
	if !errors.Is(err, ErrJWTSecretMissing) {
		t.Fatalf("expected ErrJWTSecretMissing, got %v", err)
	}
}

func TestAuthUserFromWebSocketRequest_MissingUser(t *testing.T) {
	token := signedWebSocketToken(t, &webSocketClaims{})
	req := wsRequest("/pos/ws")
	req.Header.Set("Authorization", "Bearer "+token)

	_, err := AuthUserFromWebSocketRequest(req)
	if !errors.Is(err, ErrWebSocketUserMissing) {
		t.Fatalf("expected ErrWebSocketUserMissing, got %v", err)
	}
}

func TestWebSocketTicketIsOneUseAndStoresOnlyHash(t *testing.T) {
	forceLocalWebSocketTicketStore(t)
	const rawJWT = "eyJhbGciOiJIUzI1NiJ9.secret.signature"
	ticket, err := IssueWebSocketTicket(context.Background(), &AuthUser{UserID: 11, Token: rawJWT})
	if err != nil {
		t.Fatalf("IssueWebSocketTicket: %v", err)
	}

	hash := hashWebSocketTicket(ticket)
	websocketTickets.Lock()
	_, hashStored := websocketTickets.entries[hash]
	_, rawStored := websocketTickets.entries[ticket]
	websocketTickets.Unlock()
	if !hashStored || rawStored {
		t.Fatalf("ticket store key must be the SHA-256 hash only: hash=%t raw=%t", hashStored, rawStored)
	}

	user, ok := ConsumeWebSocketTicket(context.Background(), ticket)
	if !ok || user.UserID != 11 {
		t.Fatalf("first ticket consume failed: user=%+v ok=%t", user, ok)
	}
	if _, ok := ConsumeWebSocketTicket(context.Background(), ticket); ok {
		t.Fatal("ticket was accepted more than once")
	}
}

func TestWebSocketTicketExpires(t *testing.T) {
	forceLocalWebSocketTicketStore(t)
	ticket, err := issueWebSocketTicketAt(
		context.Background(),
		&AuthUser{UserID: 12},
		time.Now().Add(-webSocketTicketLifetime-time.Second),
	)
	if err != nil {
		t.Fatalf("issueWebSocketTicketAt: %v", err)
	}

	if _, ok := ConsumeWebSocketTicket(context.Background(), ticket); ok {
		t.Fatal("expired ticket was accepted")
	}
}

// Tickets that are issued but never consumed must not accumulate in the local
// fallback map for the process lifetime.
func TestWebSocketTicketFallbackMapEvictsExpiredTickets(t *testing.T) {
	forceLocalWebSocketTicketStore(t)

	expiredTicket, err := issueWebSocketTicketAt(
		context.Background(),
		&AuthUser{UserID: 13},
		time.Now().Add(-webSocketTicketLifetime-time.Second),
	)
	if err != nil {
		t.Fatalf("issueWebSocketTicketAt: %v", err)
	}
	expiredHash := hashWebSocketTicket(expiredTicket)

	if _, err := IssueWebSocketTicket(context.Background(), &AuthUser{UserID: 14}); err != nil {
		t.Fatalf("IssueWebSocketTicket: %v", err)
	}

	websocketTickets.Lock()
	_, stillStored := websocketTickets.entries[expiredHash]
	websocketTickets.Unlock()
	if stillStored {
		t.Fatal("expired ticket remained in the fallback map after a new ticket was stored")
	}
}

func TestWebSocketTicketRedisWriteFailureIsReturned(t *testing.T) {
	err := storeWebSocketTicketWithPut(
		context.Background(),
		"ticket-that-must-not-be-issued",
		websocketTicketRecord{UserID: 15, ExpiresAt: time.Now().Add(webSocketTicketLifetime).UnixNano()},
		true,
		func(context.Context, string, []byte, time.Duration) bool { return false },
	)
	if !errors.Is(err, ErrWebSocketTicketStorage) {
		t.Fatalf("storeWebSocketTicketWithPut() error = %v, want %v", err, ErrWebSocketTicketStorage)
	}
}
