package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterRoutesIncludesContractAvailabilityPaths(t *testing.T) {
	router := chi.NewRouter()
	RegisterRoutes(context.Background(), router, nil)

	want := map[string]bool{
		"GET /hr/shift/v1/availability":         false,
		"POST /hr/shift/v1/availability":        false,
		"DELETE /hr/shift/v1/availability/{id}": false,
		"GET /hr/shifts/availability":           false,
		"POST /hr/shifts/availability":          false,
		"DELETE /hr/shifts/availability":        false,
	}
	if err := chi.Walk(router, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		key := method + " " + route
		if _, ok := want[key]; ok {
			want[key] = true
		}
		return nil
	}); err != nil {
		t.Fatalf("walk routes: %v", err)
	}
	for route, found := range want {
		if !found {
			t.Errorf("missing registered route %s", route)
		}
	}
}
