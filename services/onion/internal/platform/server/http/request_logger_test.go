package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
	if profile["name"] != "Voxire" {
		t.Fatalf("sanitizeJSON removed non-sensitive value: %s", got)
	}
}
