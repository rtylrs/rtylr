package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func TestRequestLoggerPreservesJSONBodyWithCharset(t *testing.T) {
	const body = `{"name":"Voxire"}`
	var seen string

	handler := RequestLoggerMiddleware(context.Background())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll: %v", err)
		}
		seen = string(b)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if seen != body {
		t.Fatalf("handler saw body %q, want %q", seen, body)
	}
}

func TestRequestLoggerPreservesLargeJSONBody(t *testing.T) {
	body := `{"payload":"` + strings.Repeat("x", 16<<10) + `"}`
	var seenBytes int

	handler := RequestLoggerMiddleware(context.Background())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll: %v", err)
		}
		seenBytes = len(b)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if seenBytes != len(body) {
		t.Fatalf("handler saw %d bytes, want %d", seenBytes, len(body))
	}
}

func TestSanitizeMapRedactsCredentialFieldVariants(t *testing.T) {
	for _, key := range []string{"pin", "pin_code", "cvv", "cvc", "ssn", "tax_id", "national_id", "otp", "card_number", "expiry", "passwd", "pwd", "x-api-key", "sessionid"} {
		t.Run(key, func(t *testing.T) {
			got := sanitizeMap(map[string][]string{key: {"secret"}})
			if got[key][0] != "[REDACTED]" {
				t.Fatalf("%s was not redacted: %#v", key, got)
			}
		})
	}
}

func TestDuplicateDetectorCleanupLoopRemovesStaleEntries(t *testing.T) {
	detector := newDuplicateDetector(time.Millisecond)
	detector.requests["stale"] = &DuplicateInfo{
		firstSeen:    time.Now().Add(-time.Hour),
		lastSeen:     time.Now().Add(-time.Hour),
		lastReported: time.Now().Add(-time.Hour),
		count:        1,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	detector.startCleanup(ctx, time.Millisecond)

	deadline := time.Now().Add(250 * time.Millisecond)
	for time.Now().Before(deadline) {
		detector.mu.RLock()
		_, exists := detector.requests["stale"]
		detector.mu.RUnlock()
		if !exists {
			return
		}
		time.Sleep(time.Millisecond)
	}

	t.Fatal("stale duplicate detector entry was not cleaned up")
}

func TestResponseWriterPreservesFlusher(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &ResponseWriter{ResponseWriter: rec}

	flusher, ok := any(rw).(http.Flusher)
	if !ok {
		t.Fatal("ResponseWriter should preserve http.Flusher")
	}

	flusher.Flush()
	if !rec.Flushed {
		t.Fatal("Flush was not forwarded to the wrapped response writer")
	}
}

func TestRequestRouteTemplateDoesNotExposeValuesOrQuery(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/login?token=secret", nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.RoutePatterns = []string{"/login"}
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
	if got := requestRouteTemplate(req); got != "/login" {
		t.Fatalf("requestRouteTemplate = %q, want /login", got)
	}
}
