package orders

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterRoutesIncludesDashboardOrderList(t *testing.T) {
	module := &Module{}
	router := chi.NewRouter()
	module.RegisterRoutes(context.Background(), router)

	req := httptest.NewRequest(http.MethodGet, "/pos/orders?page=1&page_size=50", http.NoBody)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code == http.StatusNotFound {
		t.Fatal("dashboard order list route was not registered")
	}
	if res.Code != http.StatusForbidden {
		t.Fatalf("anonymous request returned %d, want 403", res.Code)
	}
}
