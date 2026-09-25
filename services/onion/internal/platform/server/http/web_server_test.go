package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func testRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("pong")); err != nil {
			return
		}
	})

	return r
}

func TestComposeHandlerRunsMiddlewareChain(t *testing.T) {
	withPartialsAPIKey(t, "")

	ran := false
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ran = true
			next.ServeHTTP(w, r)
		})
	}

	rec := httptest.NewRecorder()
	composeHandler(testRouter(), []MiddlewareFunc{mw}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))

	if !ran {
		t.Fatalf("expected the middleware chain to run")
	}
	if rec.Code != http.StatusOK || rec.Body.String() != "pong" {
		t.Fatalf("expected the route to be served through the chain, got %d %q", rec.Code, rec.Body.String())
	}
}

func TestComposeHandlerBypassesMiddlewareWithValidKey(t *testing.T) {
	withPartialsAPIKey(t, "pka_test_key")

	ran := false
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ran = true
			next.ServeHTTP(w, r)
		})
	}

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(PartialsAPIKeyHeader, "pka_test_key")
	rec := httptest.NewRecorder()

	composeHandler(testRouter(), []MiddlewareFunc{mw}).ServeHTTP(rec, req)

	if ran {
		t.Fatalf("expected the %s bypass to skip the middleware chain", PartialsAPIKeyHeader)
	}
	if rec.Code != http.StatusOK || rec.Body.String() != "pong" {
		t.Fatalf("expected the bare route to be served, got %d %q", rec.Code, rec.Body.String())
	}
}

func TestComposeHandlerStillRunsMiddlewareWithWrongKey(t *testing.T) {
	withPartialsAPIKey(t, "pka_test_key")

	ran := false
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ran = true
			next.ServeHTTP(w, r)
		})
	}

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(PartialsAPIKeyHeader, "pka_wrong")
	rec := httptest.NewRecorder()

	composeHandler(testRouter(), []MiddlewareFunc{mw}).ServeHTTP(rec, req)

	if !ran {
		t.Fatalf("expected the middleware chain to run with a wrong key")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected the route to be served, got %d", rec.Code)
	}
}
