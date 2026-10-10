package workspaces

import (
	"net/http"
	"net/http/httptest"
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

	if got := (&Module{}).GetAnonymousRoutes(); len(got) != 1 || got[0] != "/flow/ws" {
		t.Fatalf("only the websocket upgrade should be anonymous, got %v", got)
	}
}
