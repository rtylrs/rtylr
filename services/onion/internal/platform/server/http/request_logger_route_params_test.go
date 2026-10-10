package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/voxire/onion/internal/platform/observability"
)

func TestRouteParamNamesFromRequestDoesNotLogValues(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /businesses/{business_id}/users/{user_id}", func(http.ResponseWriter, *http.Request) {})
	observability.SetRouteMux(mux)
	t.Cleanup(func() { observability.SetRouteMux(nil) })

	req := httptest.NewRequest("GET", "/businesses/42/users/7", nil)

	got := routeParamNamesFromRequest(req)
	if len(got) != 2 {
		t.Fatalf("len(routeParams) = %d, want 2", len(got))
	}
	if got[0] != "business_id" {
		t.Fatalf("first route param = %q, want %q", got[0], "business_id")
	}
	if got[1] != "user_id" {
		t.Fatalf("second route param = %q, want %q", got[1], "user_id")
	}
}
