package http

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestBodyLimitExemptPaths(t *testing.T) {
	t.Parallel()

	exempt := []string{
		"/careers/apply",
		"/voxhire/applications",
		"/voxhire/careers/apply",
		"/upload",
	}
	for _, path := range exempt {
		if !requestBodyLimitExempt(path) {
			t.Fatalf("expected %q to be exempt from global body limit", path)
		}
	}

	if requestBodyLimitExempt("/auth/login") {
		t.Fatal("expected /auth/login to be subject to global body limit")
	}
}

func TestRequestBodyLimitMiddlewareRejectsOversizedBody(t *testing.T) {
	t.Parallel()

	const limit = 32
	handler := RequestBodyLimitMiddleware(limit)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			if isRequestBodyTooLarge(err) {
				writeRequestEntityTooLarge(r.Context(), w)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	body := bytes.Repeat([]byte("x"), limit+1)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d body %q", rec.Code, http.StatusRequestEntityTooLarge, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "too large") {
		t.Fatalf("expected too large message, got %q", rec.Body.String())
	}
}

func TestRequestBodyLimitMiddlewareSkipsExemptPath(t *testing.T) {
	t.Parallel()

	const limit = 32
	handler := RequestBodyLimitMiddleware(limit)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if len(data) <= limit {
			t.Fatalf("expected body larger than %d on exempt path, got %d", limit, len(data))
		}
		w.WriteHeader(http.StatusOK)
	}))

	body := bytes.Repeat([]byte("x"), limit+1)
	req := httptest.NewRequest(http.MethodPost, "/careers/apply", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}
