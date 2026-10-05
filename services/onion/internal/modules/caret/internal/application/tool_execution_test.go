package application

import (
	"context"
	"errors"
	"testing"
)

func TestToolCallExecutionStateRejectsDuplicatesAndRunawayLoops(t *testing.T) {
	state := newToolCallExecutionState()
	input := map[string]any{"limit": 5}
	if err := state.admit("crm.list_customers", input); err != nil {
		t.Fatalf("first call rejected: %v", err)
	}
	if !errors.Is(state.admit("crm.list_customers", input), ErrDuplicateToolCall) {
		t.Fatal("expected duplicate tool call to be rejected")
	}
	for i := state.count; i < MaxToolCallsPerTurn; i++ {
		if err := state.admit("crm.list_customers", map[string]any{"limit": i + 100}); err != nil {
			t.Fatalf("unique call %d rejected: %v", i, err)
		}
	}
	if !errors.Is(state.admit("crm.list_customers", map[string]any{"limit": 999}), ErrToolCallLimit) {
		t.Fatal("expected per-turn tool call limit")
	}
}

func TestValidateToolInvocationBoundsInputsAndMode(t *testing.T) {
	tool, ok := ToolByName("crm.list_customers")
	if !ok {
		t.Fatal("crm.list_customers missing from registry")
	}
	if err := ValidateToolInvocation(ModeAsk, tool, map[string]any{"limit": 10}); err != nil {
		t.Fatalf("valid read input rejected: %v", err)
	}
	if err := ValidateToolInvocation(ModeSupport, tool, nil); err == nil {
		t.Fatal("support mode should not invoke business tools")
	}
	if err := ValidateToolInvocation(ModeAsk, tool, map[string]any{"payload": string(make([]byte, MaxToolStringValueLen+1))}); err == nil {
		t.Fatal("oversized string input should be rejected")
	}
}

func TestExecuteToolWithTimeoutStopsAtFirstMatchingExecutor(t *testing.T) {
	first := recordingToolExecutor{output: map[string]any{"total": 3}, executed: true}
	second := recordingToolExecutor{output: map[string]any{"total": 9}, executed: true}
	output, executed, duration, err := ExecuteToolWithTimeout(context.Background(), []ToolExecutor{first, second}, "crm.list_customers", nil)
	if err != nil || !executed {
		t.Fatalf("executeToolWithTimeout = output=%#v executed=%v err=%v", output, executed, err)
	}
	if output["total"] != 3 {
		t.Fatalf("first matching executor was not used: %#v", output)
	}
	if duration <= 0 {
		t.Fatal("expected execution duration")
	}
}

func TestExecuteToolWithTimeoutRecoversExecutorPanic(t *testing.T) {
	output, executed, _, err := ExecuteToolWithTimeout(context.Background(), []ToolExecutor{panicToolExecutor{}}, "crm.list_customers", nil)
	if err == nil || executed || output != nil {
		t.Fatalf("panic should become an error: output=%#v executed=%v err=%v", output, executed, err)
	}
}

type recordingToolExecutor struct {
	output   map[string]any
	executed bool
}

func (f recordingToolExecutor) ExecuteCaretTool(context.Context, string, map[string]any) (map[string]any, bool, error) {
	return f.output, f.executed, nil
}

type panicToolExecutor struct{}

func (panicToolExecutor) ExecuteCaretTool(context.Context, string, map[string]any) (map[string]any, bool, error) {
	panic("test executor panic")
}

func TestExecuteToolWithTimeoutHonorsCanceledParentContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, _, err := ExecuteToolWithTimeout(ctx, []ToolExecutor{recordingToolExecutor{}}, "crm.list_customers", nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled context, got %v", err)
	}
}
