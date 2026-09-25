package erpsettings

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterAPIRegistersERPContractRoutes(t *testing.T) {
	mux := http.NewServeMux()
	NewModule(context.Background(), nil).RegisterAPI(mux)

	for _, route := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/erp/settings/v1/settings"},
		{method: http.MethodPut, path: "/erp/settings/v1/settings"},
	} {
		t.Run(route.method, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, http.NoBody)
			_, pattern := mux.Handler(req)
			want := route.method + " " + route.path
			if pattern != want {
				t.Fatalf("route pattern = %q, want %q", pattern, want)
			}
		})
	}
}
