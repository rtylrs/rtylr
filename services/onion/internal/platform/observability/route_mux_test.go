package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRequestRouteTemplateAsksTheSharedMuxBehindTheCatchAll(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /crm/sales/leads/{id}/touchpoints", func(http.ResponseWriter, *http.Request) {})
	SetRouteMux(mux)
	t.Cleanup(func() { SetRouteMux(nil) })

	req := httptest.NewRequest(http.MethodGet, "/crm/sales/leads/42/touchpoints?token=secret", nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.RoutePatterns = []string{"/*"}
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	if got := RequestRouteTemplate(req); got != "/crm/sales/leads/{id}/touchpoints" {
		t.Fatalf("template = %q", got)
	}
	if got := RouteParamNames("/crm/sales/leads/{id}/touchpoints/{rest...}"); !reflect.DeepEqual(got, []string{"id", "rest"}) {
		t.Fatalf("param names = %v", got)
	}
}
