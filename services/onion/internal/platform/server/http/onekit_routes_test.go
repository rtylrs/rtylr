package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestMountOneKitStripsPublicAPIPrefix(t *testing.T) {
	oneKitMux := http.NewServeMux()
	oneKitMux.HandleFunc("GET /hub/user/v1/users/me", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	router := chi.NewRouter()
	MountOneKit(router, oneKitMux)

	request := httptest.NewRequest(http.MethodGet, "/api/hub/user/v1/users/me", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected generated route to return %d, got %d", http.StatusNoContent, response.Code)
	}
}
