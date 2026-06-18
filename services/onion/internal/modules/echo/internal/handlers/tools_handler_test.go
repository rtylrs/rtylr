package handlers

import (
	"context"
	"net/http"
	"testing"
)

func TestToolDecisionValidateRequiresKnownDecision(t *testing.T) {
	handler := &ToolDecisionHandler{}

	err := handler.Validate(context.Background(), &ToolDecisionRequest{
		ToolCallID: "tool-1",
		Name:       "crm.create_customer",
		Decision:   "maybe",
	})

	if err == nil {
		t.Fatal("expected validation error")
	}
	if err.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %#v", err)
	}
}

func TestToolDecisionDenyReturnsDeniedWithoutExecution(t *testing.T) {
	handler := &ToolDecisionHandler{}

	got, err := handler.Serve(context.Background(), &ToolDecisionRequest{
		ToolCallID: "tool-1",
		Name:       "crm.create_customer",
		Input:      map[string]any{"name": "Ada"},
		Decision:   "deny",
	})

	if err != nil {
		t.Fatalf("Serve returned error: %#v", err)
	}
	if got.Status != "denied" {
		t.Fatalf("expected denied status, got %#v", got)
	}
	if got.Executed {
		t.Fatalf("denied tool must not execute: %#v", got)
	}
	if got.Tool.Name != "crm.create_customer" {
		t.Fatalf("expected registered tool metadata, got %#v", got.Tool)
	}
}

func TestToolDecisionApproveUnimplementedWriteToolReturnsExplicitError(t *testing.T) {
	handler := &ToolDecisionHandler{}

	got, err := handler.Serve(context.Background(), &ToolDecisionRequest{
		ToolCallID: "tool-1",
		Name:       "crm.create_customer",
		Input:      map[string]any{"name": "Ada"},
		Decision:   "approve",
	})

	if err != nil {
		t.Fatalf("Serve returned error: %#v", err)
	}
	if got.Status != "error" {
		t.Fatalf("expected explicit error status, got %#v", got)
	}
	if got.Executed {
		t.Fatalf("unimplemented write tool must not execute: %#v", got)
	}
	if got.Message == "" {
		t.Fatalf("expected user-facing message, got %#v", got)
	}
}

func TestToolDecisionUnknownToolReturnsExplicitError(t *testing.T) {
	handler := &ToolDecisionHandler{}

	got, err := handler.Serve(context.Background(), &ToolDecisionRequest{
		ToolCallID: "tool-1",
		Name:       "crm.destroy_everything",
		Decision:   "approve",
	})

	if err != nil {
		t.Fatalf("Serve returned error: %#v", err)
	}
	if got.Status != "error" {
		t.Fatalf("expected explicit error status, got %#v", got)
	}
	if got.Executed {
		t.Fatalf("unknown tool must not execute: %#v", got)
	}
}
