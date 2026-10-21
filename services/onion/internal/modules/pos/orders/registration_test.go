package orders

import (
	"context"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterRoutesIncludesDashboardOrderList(t *testing.T) {
	module := &Module{}
	router := serverHttp.NewLegacyRoutes()
	module.RegisterRoutes(context.Background(), router)

	req := httptest.NewRequest(http.MethodGet, "/pos/orders?page=1&page_size=50", http.NoBody)
	res := httptest.NewRecorder()
	router.Handler().ServeHTTP(res, req)

	if res.Code == http.StatusNotFound {
		t.Fatal("dashboard order list route was not registered")
	}
	if res.Code != http.StatusForbidden {
		t.Fatalf("anonymous request returned %d, want 403", res.Code)
	}
}
