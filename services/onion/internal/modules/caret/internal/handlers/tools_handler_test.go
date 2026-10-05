package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	echoapp "github.com/voxire/onion/internal/modules/caret/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
)

func oneKitStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func TestToolDecisionRequiresAuthenticatedPendingApprovalInProductionMode(t *testing.T) {
	got, err := (&ToolDecisionHandler{RequirePendingApproval: true}).Serve(context.Background(), &ToolDecisionRequest{
		ToolCallID: "tool-auth-required",
		Name:       "flow.create_task",
		Input:      map[string]any{"title": "Follow up"},
		Decision:   "approve",
		Mode:       string(ModeAgent),
	})
	if err != nil {
		t.Fatalf("Serve returned error: %v", err)
	}
	if got.Status != "error" || !strings.Contains(got.Message, "Authentication") {
		t.Fatalf("response = %#v", got)
	}
}

func TestPendingApprovalCannotBeReplayed(t *testing.T) {
	ctx := context.WithValue(context.Background(), authPkg.AuthUserKey, &authPkg.AuthUser{UserID: 7, BusinessID: 9})
	input := map[string]any{"title": "Follow up"}
	tool, _ := echoapp.ToolByName("flow.create_task")
	if err := echoapp.RegisterPendingToolApproval(ctx, "tool-replay", tool, input, echoapp.ModeAgent); err != nil {
		t.Fatalf("registerPendingToolApproval: %v", err)
	}
	handler := &ToolDecisionHandler{RequirePendingApproval: true, Executor: fakeToolExecutor{output: map[string]any{"task": map[string]any{"id": 1}}}}
	req := &ToolDecisionRequest{ToolCallID: "tool-replay", Name: "flow.create_task", Input: input, Decision: "approve", Mode: string(ModeAgent)}
	first, err := handler.Serve(ctx, req)
	if err != nil || first.Status != "success" {
		t.Fatalf("first approval = %#v, err=%v", first, err)
	}
	second, err := handler.Serve(ctx, req)
	if err != nil || second.Status != "error" || second.Executed {
		t.Fatalf("replayed approval = %#v, err=%v", second, err)
	}
}

func TestListToolsHandlerHonorsModeCapability(t *testing.T) {
	handler := &ListToolsHandler{}

	for _, mode := range []string{string(ModeAsk), string(ModeAnalyst)} {
		got, err := handler.Serve(context.Background(), &ListToolsRequest{Mode: mode})
		if err != nil {
			t.Fatalf("Serve returned error: %#v", err)
		}
		if len(got.Tools) == 0 {
			t.Fatalf("%s mode should include read-only tools", mode)
		}
		for _, tool := range got.Tools {
			if tool.Scope != "read" {
				t.Fatalf("%s mode exposed write tool %s", mode, tool.Name)
			}
		}
	}

	agent, err := handler.Serve(context.Background(), &ListToolsRequest{Mode: string(ModeAgent)})
	if err != nil {
		t.Fatalf("Serve returned error: %#v", err)
	}
	foundWrite := false
	for _, tool := range agent.Tools {
		if tool.Scope == "write" {
			foundWrite = true
			break
		}
	}
	if !foundWrite {
		t.Fatalf("agent mode should include write-capable tools")
	}

	support, err := handler.Serve(context.Background(), &ListToolsRequest{Mode: string(ModeSupport)})
	if err != nil {
		t.Fatalf("Serve returned error: %#v", err)
	}
	if len(support.Tools) != 0 {
		t.Fatalf("support mode should not include tools: %#v", support.Tools)
	}
}

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
	if oneKitStatus(err) != http.StatusBadRequest {
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
	if got.Preview == "" || !strings.Contains(got.Preview, "create a customer") {
		t.Fatalf("expected friendly action preview, got %#v", got)
	}
}

func TestToolDecisionApproveUnimplementedWriteToolReturnsExplicitError(t *testing.T) {
	handler := &ToolDecisionHandler{}

	got, err := handler.Serve(context.Background(), &ToolDecisionRequest{
		ToolCallID: "tool-1",
		Name:       "crm.create_customer",
		Input:      map[string]any{"name": "Ada"},
		Decision:   "approve",
		Mode:       string(ModeAgent),
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

func TestToolDecisionApproveWriteToolRequiresAgentMode(t *testing.T) {
	handler := &ToolDecisionHandler{
		Executor: fakeToolExecutor{
			output: map[string]any{"customer": map[string]any{"id": int64(42), "first_name": "Ada"}},
		},
	}

	got, err := handler.Serve(context.Background(), &ToolDecisionRequest{
		ToolCallID: "tool-1",
		Name:       "crm.create_customer",
		Input:      map[string]any{"first_name": "Ada"},
		Decision:   "approve",
		Mode:       string(ModeAsk),
	})
	if err != nil {
		t.Fatalf("Serve returned error: %#v", err)
	}
	if got.Status != "error" || got.Executed {
		t.Fatalf("expected mode rejection without execution, got %#v", got)
	}
	if !strings.Contains(got.Message, "Agent mode") {
		t.Fatalf("expected mode guidance, got %#v", got.Message)
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
		Mode:       string(ModeAgent),
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
		Mode:       string(ModeAgent),
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
		Mode:       string(ModeAgent),
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
