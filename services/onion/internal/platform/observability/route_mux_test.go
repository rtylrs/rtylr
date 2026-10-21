package observability

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestRequestRouteTemplateResolvesThroughTheServerMux(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /crm/sales/leads/{id}/touchpoints", func(http.ResponseWriter, *http.Request) {})
	SetRouteMux(mux)
	t.Cleanup(func() { SetRouteMux(nil) })

	req := httptest.NewRequest(http.MethodGet, "/crm/sales/leads/42/touchpoints?token=secret", nil)
	if got := RequestRouteTemplate(req); got != "/crm/sales/leads/{id}/touchpoints" {
		t.Fatalf("template = %q", got)
	}
	if got := RequestRouteTemplate(httptest.NewRequest(http.MethodGet, "/nowhere", nil)); got != "[unmatched route]" {
		t.Fatalf("unmatched template = %q", got)
	}
	if got := RouteParamNames("/crm/sales/leads/{id}/touchpoints/{rest...}"); !reflect.DeepEqual(got, []string{"id", "rest"}) {
		t.Fatalf("param names = %v", got)
	}
}
