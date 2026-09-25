package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/platform/config"
)

func TestInvalidateCacheRequiresInternalToken(t *testing.T) {
	withInternalToken(t, "secret")

	router := chi.NewRouter()
	RegisterCacheRoutes(context.Background(), router)

	req := httptest.NewRequest(http.MethodPost, "/invalidate-cache", strings.NewReader(`{"type":"all"}`))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden without Internal-Token, got %d", rec.Code)
	}
}

func TestInvalidateCacheAcceptsInternalToken(t *testing.T) {
	withInternalToken(t, "secret")

	router := chi.NewRouter()
	RegisterCacheRoutes(context.Background(), router)

	req := httptest.NewRequest(http.MethodPost, "/invalidate-cache", strings.NewReader(`{"type":"all"}`))
	req.Header.Set("Internal-Token", "secret")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected OK with Internal-Token, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestInvalidateCacheRejectsTrailingJSON(t *testing.T) {
	withInternalToken(t, "secret")

	router := chi.NewRouter()
	RegisterCacheRoutes(context.Background(), router)

	req := httptest.NewRequest(http.MethodPost, "/invalidate-cache", strings.NewReader(`{"type":"all"}{"type":"ids"}`))
	req.Header.Set("Internal-Token", "secret")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request for trailing JSON, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestInvalidateCacheRejectsRequestsThatWouldBeSkipped(t *testing.T) {
	withInternalToken(t, "secret")

	for _, body := range []string{
		`{"type":"user"}`,
		`{"type":"user","ids":["user-42"]}`,
		`{"type":"user","ids":["42 "]}`,
		`{"type":"team","ids":["business-42"]}`,
	} {
		t.Run(body, func(t *testing.T) {
			router := chi.NewRouter()
			RegisterCacheRoutes(context.Background(), router)

			req := httptest.NewRequest(http.MethodPost, "/invalidate-cache", strings.NewReader(body))
			req.Header.Set("Internal-Token", "secret")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected bad request for %s, got %d: %s", body, rec.Code, rec.Body.String())
			}
		})
	}
}

func withInternalToken(t *testing.T, token string) {
	t.Helper()
	old := config.GlobalConfig
	config.GlobalConfig = &config.AppConfig{InternalToken: token}
	t.Cleanup(func() { config.GlobalConfig = old })
}
