package http

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRouteParamNamesFromRequestDoesNotLogValues(t *testing.T) {
	req := httptest.NewRequest("GET", "/businesses/42/users/7", nil)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("business_id", "42")
	routeCtx.URLParams.Add("user_id", "7")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

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
