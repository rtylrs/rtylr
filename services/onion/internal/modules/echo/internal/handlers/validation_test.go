package handlers

import (
	"context"
	"net/http"
	"testing"
)

func TestCaretHandlersRejectNilRequests(t *testing.T) {
	if err := (&GenerateTextStreamHandler{}).Validate(context.Background(), nil); oneKitStatus(err) != http.StatusBadRequest {
		t.Fatalf("stream nil validation = %v", err)
	}
	if err := (&ToolDecisionHandler{}).Validate(context.Background(), nil); oneKitStatus(err) != http.StatusBadRequest {
		t.Fatalf("tool decision nil validation = %v", err)
	}
	if _, err := (&ListToolsHandler{}).Serve(context.Background(), nil); err != nil {
		t.Fatalf("nil list-tools request should use default mode: %v", err)
	}
}
