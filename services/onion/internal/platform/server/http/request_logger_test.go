package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

func TestRequestLoggerPreservesLargeJSONBodyAfterPreviewLimit(t *testing.T) {
	body := `{"payload":"` + strings.Repeat("x", maxLoggedRequestBodyBytes*2) + `"}`
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

func TestSanitizeJSONTruncatesLargePayloads(t *testing.T) {
	large := `{"items":[` + strings.Repeat(`"x",`, 500) + `"end"]}`
	got := sanitizeJSON(large)
	if len(got) > maxLoggedBodyPreviewBytes {
		t.Fatalf("sanitized body length %d exceeds preview cap %d", len(got), maxLoggedBodyPreviewBytes)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("expected truncated preview suffix, got %q", got)
	}
}

func TestSanitizeJSONRedactsSensitiveFieldVariants(t *testing.T) {
	got := sanitizeJSON(`{
		"password_hash":"hash",
		"access_token":"token",
		"profile":{"api_key":"key","name":"Voxire"}
	}`)

	var decoded map[string]any
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("sanitized body should still be JSON: %v", err)
	}
	if decoded["password_hash"] != "[REDACTED]" {
		t.Fatalf("password_hash was not redacted: %s", got)
	}
	if decoded["access_token"] != "[REDACTED]" {
		t.Fatalf("access_token was not redacted: %s", got)
	}
	profile, ok := decoded["profile"].(map[string]any)
	if !ok {
		t.Fatalf("profile is missing or invalid: %s", got)
	}
	if profile["api_key"] != "[REDACTED]" {
		t.Fatalf("nested api_key was not redacted: %s", got)
	}
	if profile["name"] != "[REDACTED]" {
		t.Fatalf("name was not redacted: %s", got)
	}
}

func TestSanitizeJSONRedactsPersonalData(t *testing.T) {
	got := sanitizeJSON(`{"email":"person@example.com","phone_number":"+15555550123","business_name":"Example Shop"}`)
	for _, value := range []string{"person@example.com", "+15555550123", "Example Shop"} {
		if strings.Contains(got, value) {
			t.Fatalf("personal value %q leaked into request log: %s", value, got)
		}
	}
}

func TestSanitizeJSONDoesNotEchoMalformedSensitiveBody(t *testing.T) {
	got := sanitizeJSON(`{"password":"secret"`)
	if strings.Contains(got, "secret") || strings.Contains(got, "password") {
		t.Fatalf("malformed body leaked sensitive content: %q", got)
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

func TestRequestURLWithoutQuery(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/login?token=secret", nil)
	if got := requestURLWithoutQuery(req); got != "/login" {
		t.Fatalf("requestURLWithoutQuery = %q, want /login", got)
	}
}
