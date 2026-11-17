package workspaces

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterAPIIncludesProtectedRealtimeTicketEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	(&Module{}).RegisterAPI(mux)

	req := httptest.NewRequest(http.MethodPost, "/flow/ws-ticket", http.NoBody)
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)

	// The route itself is registered, but the handler must reject a request
	// that has not passed the normal Bearer middleware.
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected registered ticket route to return 401 without auth, got %d", res.Code)
	}

	// The websocket upgrade route is registered too (a plain GET without the
	// upgrade handshake or a ticket is refused, not 404).
	ws := httptest.NewRecorder()
	mux.ServeHTTP(ws, httptest.NewRequest(http.MethodGet, "/flow/ws", http.NoBody))
	if ws.Code == http.StatusNotFound || ws.Code == http.StatusMethodNotAllowed {
		t.Fatalf("expected the websocket route to be registered, got %d", ws.Code)
	}

	// The GitHub webhook is registered and, with no webhook secret, refuses
	// every delivery instead of accepting unsigned ones.
	hook := httptest.NewRecorder()
	mux.ServeHTTP(hook, httptest.NewRequest(http.MethodPost, "/flow/github/webhook", strings.NewReader("{}")))
	if hook.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected the GitHub webhook to refuse deliveries without a secret, got %d", hook.Code)
	}

	got := (&Module{}).GetAnonymousRoutes()
	if len(got) != 2 || got[0] != "/flow/ws" || got[1] != "/flow/github/webhook" {
		t.Fatalf("only the websocket upgrade and the signed GitHub webhook should be anonymous, got %v", got)
	}
}
