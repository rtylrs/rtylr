package handlers

import (
	"context"
	"errors"
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

func TestToolDecisionApproveExecutableWriteToolReturnsSuccess(t *testing.T) {
	handler := &ToolDecisionHandler{
		Executor: fakeToolExecutor{
			output: map[string]any{"task": map[string]any{"id": int64(123), "title": "Follow up"}},
		},
	}

	got, err := handler.Serve(context.Background(), &ToolDecisionRequest{
		ToolCallID: "tool-1",
		Name:       "flow.create_task",
		Input:      map[string]any{"title": "Follow up"},
		Decision:   "approve",
	})

	if err != nil {
		t.Fatalf("Serve returned error: %#v", err)
	}
	if got.Status != "success" || !got.Executed {
		t.Fatalf("expected executed success, got %#v", got)
	}
	if got.Tool.Name != "flow.create_task" {
		t.Fatalf("expected registered Flow tool metadata, got %#v", got.Tool)
	}
	if got.Output["task"] == nil {
		t.Fatalf("expected task output, got %#v", got.Output)
	}
}

func TestToolDecisionApproveTriesMultipleExecutors(t *testing.T) {
	handler := &ToolDecisionHandler{
		Executors: []ToolExecutor{
			fakeToolExecutor{executed: false},
			fakeToolExecutor{
				output: map[string]any{"customer": map[string]any{"id": int64(42), "first_name": "Ada"}},
			},
		},
	}

	got, err := handler.Serve(context.Background(), &ToolDecisionRequest{
		ToolCallID: "tool-1",
		Name:       "crm.create_customer",
		Input:      map[string]any{"first_name": "Ada"},
		Decision:   "approve",
	})

	if err != nil {
		t.Fatalf("Serve returned error: %#v", err)
	}
	if got.Status != "success" || !got.Executed {
		t.Fatalf("expected second executor success, got %#v", got)
	}
	if got.Output["customer"] == nil {
		t.Fatalf("expected customer output, got %#v", got.Output)
	}
}

func TestToolDecisionApproveExecutableWriteToolFailureReturnsError(t *testing.T) {
	handler := &ToolDecisionHandler{
		Executor: fakeToolExecutor{err: errors.New("title is required")},
	}

	got, err := handler.Serve(context.Background(), &ToolDecisionRequest{
		ToolCallID: "tool-1",
		Name:       "flow.create_task",
		Decision:   "approve",
	})

	if err != nil {
		t.Fatalf("Serve returned error: %#v", err)
	}
	if got.Status != "error" || got.Executed {
		t.Fatalf("expected non-executed error, got %#v", got)
	}
	if got.Output["error"] != "title is required" {
		t.Fatalf("expected user-facing error output, got %#v", got.Output)
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

type fakeToolExecutor struct {
	output   map[string]any
	executed bool
	err      error
}

func (f fakeToolExecutor) ExecuteCaretTool(_ context.Context, _ string, _ map[string]any) (map[string]any, bool, error) {
	if f.err != nil {
		return nil, false, f.err
	}
	executed := f.executed
	if f.output != nil {
		executed = true
	}
	return f.output, executed, nil
}
