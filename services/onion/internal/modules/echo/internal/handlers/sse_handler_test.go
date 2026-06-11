package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateTextStreamParseRejectsTrailingJSON(t *testing.T) {
	handler := &GenerateTextStreamHandler{}
	req := httptest.NewRequest(
		http.MethodPost,
		"/ai/generate-text-stream",
		strings.NewReader(`{"prompt":"hello","user_id":1}{"prompt":"ignored","user_id":2}`),
	)
	rec := httptest.NewRecorder()

	got, err := handler.parseAndValidateRequest(rec, req)

	if !errors.Is(err, ErrInvalidRequestBody) {
		t.Fatalf("expected invalid request body, got req=%v err=%v", got, err)
	}
}

func TestGenerateTextStreamParseAcceptsCamelCaseAliases(t *testing.T) {
	handler := &GenerateTextStreamHandler{}
	req := httptest.NewRequest(
		http.MethodPost,
		"/ai/generate-text-stream",
		strings.NewReader(`{"prompt":"hello","sessionId":"session-1","userId":7}`),
	)
	rec := httptest.NewRecorder()

	got, err := handler.parseAndValidateRequest(rec, req)

	if err != nil {
		t.Fatalf("parseAndValidateRequest: %v", err)
	}
	if got.SessionID != "session-1" {
		t.Fatalf("expected session alias to map, got %q", got.SessionID)
	}
	if got.UserID != 7 {
		t.Fatalf("expected user alias to map, got %d", got.UserID)
	}
}
