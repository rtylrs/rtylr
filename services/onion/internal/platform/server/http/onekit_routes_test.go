package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestMountOneKitUsesGeneratedContractPath(t *testing.T) {
	oneKitMux := http.NewServeMux()
	oneKitMux.HandleFunc("GET /hub/user/v1/users/me", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	router := chi.NewRouter()
	MountOneKit(router, oneKitMux)

	request := httptest.NewRequest(http.MethodGet, "/hub/user/v1/users/me", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected generated route to return %d, got %d", http.StatusNoContent, response.Code)
	}
}

func TestMountOneKitDoesNotShadowPlatformRoutes(t *testing.T) {
	oneKitMux := http.NewServeMux()
	oneKitMux.HandleFunc("GET /hub/user/v1/users/me", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	router := chi.NewRouter()
	MountOneKit(router, oneKitMux)
	router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected platform route to return %d, got %d", http.StatusOK, response.Code)
	}
}
