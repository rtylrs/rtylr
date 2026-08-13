package workspaces

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterRoutesIncludesProtectedRealtimeTicketEndpoint(t *testing.T) {
	router := chi.NewRouter()
	(&Module{}).RegisterRoutes(context.Background(), router)

	req := httptest.NewRequest(http.MethodPost, "/flow/ws-ticket", http.NoBody)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	// The route itself is registered, but the handler must reject a request
	// that has not passed the normal Bearer middleware.
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected registered ticket route to return 401 without auth, got %d", res.Code)
	}

	if got := (&Module{}).GetAnonymousRoutes(); len(got) != 1 || got[0] != "/flow/ws" {
		t.Fatalf("only the websocket upgrade should be anonymous, got %v", got)
	}
}
