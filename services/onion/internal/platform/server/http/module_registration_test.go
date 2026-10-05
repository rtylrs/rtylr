package http

import (
	"context"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

type testAPIRegistrar struct{}

func (testAPIRegistrar) RegisterAPI(mux *stdhttp.ServeMux) {
	mux.HandleFunc("GET /registered", func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusNoContent)
	})
}

type testRouteRegistrar struct{}

func (testRouteRegistrar) RegisterRoutes(_ context.Context, router *chi.Mux, _ ...interface{}) {
	router.Get("/registered", func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusNoContent)
	})
}

func TestAutoRegisterAPIs(t *testing.T) {
	mux := stdhttp.NewServeMux()
	AutoRegisterAPIs(context.Background(), mux, testAPIRegistrar{})
	assertNoContent(t, mux, "/registered")
}

func TestAutoRegisterRoutes(t *testing.T) {
	router := chi.NewRouter()
	AutoRegisterRoutes(context.Background(), router, testRouteRegistrar{})
	assertNoContent(t, router, "/registered")
}

func assertNoContent(t *testing.T, handler stdhttp.Handler, path string) {
	t.Helper()
	request := httptest.NewRequest(stdhttp.MethodGet, path, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != stdhttp.StatusNoContent {
		t.Fatalf("expected %d, got %d", stdhttp.StatusNoContent, response.Code)
	}
}
