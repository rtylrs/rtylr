package handlers

import (
	"context"
	"time"

	echoapp "github.com/voxire/onion/internal/modules/echo/internal/application"
)

type ToolExecutor = echoapp.ToolExecutor
type toolCallExecutionState = echoapp.ToolCallExecutionState

func newToolCallExecutionState() *toolCallExecutionState { return echoapp.NewToolCallExecutionState() }

const (
	maxToolCallsPerTurn   = echoapp.MaxToolCallsPerTurn
	maxToolStringValueLen = echoapp.MaxToolStringValueLen
	maxToolOutputBytes    = echoapp.MaxToolOutputBytes
)

var (
	ErrToolCallLimit     = echoapp.ErrToolCallLimit
	ErrDuplicateToolCall = echoapp.ErrDuplicateToolCall
)

func toolCallFingerprint(name string, input map[string]any) string {
	return echoapp.ToolCallFingerprint(name, input)
}
func validateToolInvocation(mode CaretMode, tool ToolDef, input map[string]any) error {
	return echoapp.ValidateToolInvocation(mode, tool, input)
}
func executeToolWithTimeout(ctx context.Context, executors []ToolExecutor, name string, input map[string]any) (map[string]any, bool, error, time.Duration) {
	return echoapp.ExecuteToolWithTimeout(ctx, executors, name, input)
}
func normalizeToolInput(name string, input map[string]any) map[string]any {
	return echoapp.NormalizeToolInput(name, input)
}
func sanitizeToolOutput(output map[string]any) map[string]any {
	return echoapp.SanitizeToolOutput(output)
}
func registerPendingToolApproval(ctx context.Context, id string, tool ToolDef, input map[string]any, mode CaretMode) error {
	return echoapp.RegisterPendingToolApproval(ctx, id, tool, input, mode)
}
func consumePendingToolApproval(ctx context.Context, id string, tool ToolDef, input map[string]any, mode CaretMode) error {
	return echoapp.ConsumePendingToolApproval(ctx, id, tool, input, mode)
}
func authenticatedIdentity(ctx context.Context) (int64, int64) {
	return echoapp.AuthenticatedIdentity(ctx)
}
