package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	echoapp "github.com/voxire/onion/internal/modules/caret/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

// ListToolsRoutes is the HTTP path for the tool registry endpoint.
func ListToolsRoutes() []string { return []string{"/caret/tools"} }

// ToolDecisionRoutes is the HTTP path for user decisions on model-proposed
// tool calls.
func ToolDecisionRoutes() []string { return []string{"/caret/tools/decision"} }

// ListToolsHandler returns the tools available to the calling user. Today the
// list is a static, mode-agnostic seed registry — the eventual plan is to
// gate tools by the caller's role + active plan.
type ListToolsHandler struct{}

type ToolDecisionHandler struct {
	Executor               ToolExecutor
	Executors              []ToolExecutor
	RequirePendingApproval bool
}

type ListToolsRequest struct {
	Mode string `json:"mode"`
}

type ListToolsResponse struct {
	Tools []ToolDef `json:"tools"`
}

type ToolDecisionRequest struct {
	ToolCallID string         `json:"tool_call_id"`
	Name       string         `json:"name"`
	Input      map[string]any `json:"input"`
	Decision   string         `json:"decision"`
	Mode       string         `json:"mode"`
}

type ToolDecisionResponse struct {
	ToolCallID string         `json:"tool_call_id"`
	Status     string         `json:"status"`
	Message    string         `json:"message"`
	Executed   bool           `json:"executed"`
	Tool       ToolDef        `json:"tool"`
	Preview    string         `json:"preview,omitempty"`
	Output     map[string]any `json:"output,omitempty"`
	DurationMs int64          `json:"duration_ms,omitempty"`
}

func (h *ListToolsHandler) Serve(_ context.Context, req *ListToolsRequest) (ListToolsResponse, error) {
	if req == nil {
		req = &ListToolsRequest{}
	}
	mode := echoapp.ParseMode(req.Mode)
	return ListToolsResponse{Tools: echoapp.ToolsForMode(mode)}, nil
}

func (h *ToolDecisionHandler) Validate(_ context.Context, req *ToolDecisionRequest) error {
	if req == nil {
		return toolDecisionValidationError("request is required")
	}
	if strings.TrimSpace(req.ToolCallID) == "" {
		return toolDecisionValidationError("tool_call_id is required")
	}
	if len(strings.TrimSpace(req.ToolCallID)) > 128 {
		return toolDecisionValidationError("tool_call_id is too long")
	}
	if strings.TrimSpace(req.Name) == "" {
		return toolDecisionValidationError("name is required")
	}
	switch strings.TrimSpace(strings.ToLower(req.Decision)) {
	case "approve", "deny":
		return nil
	default:
		return toolDecisionValidationError("decision must be approve or deny")
	}
}

func (h *ToolDecisionHandler) Serve(ctx context.Context, req *ToolDecisionRequest) (ToolDecisionResponse, error) {
	decision := strings.TrimSpace(strings.ToLower(req.Decision))
	name := strings.TrimSpace(req.Name)
	tool, ok := echoapp.ToolByName(name)
	if !ok {
		logger.Ctx(ctx).Warn("caret tool decision rejected for unknown tool",
			logger.String("tool_call_id", req.ToolCallID),
			logger.String("tool_name", name),
			logger.String("decision", decision),
		)

		return ToolDecisionResponse{
			ToolCallID: req.ToolCallID,
			Status:     "error",
			Message:    "Tool is not registered. No changes were made.",
			Executed:   false,
		}, nil
	}
	if h.RequirePendingApproval {
		businessID, userID := echoapp.AuthenticatedIdentity(ctx)
		if businessID <= 0 || userID <= 0 {
			return ToolDecisionResponse{ToolCallID: req.ToolCallID, Status: "error", Message: "Authentication is required to approve or deny this action.", Tool: tool}, nil
		}
	}

	if decision == "deny" {
		if h.RequirePendingApproval {
			if err := echoapp.ConsumePendingToolApproval(ctx, req.ToolCallID, tool, req.Input, ""); err != nil {
				return ToolDecisionResponse{ToolCallID: req.ToolCallID, Status: "error", Message: "This tool proposal is no longer valid. No changes were made.", Tool: tool}, nil
			}
		}
		h.logDecision(ctx, req, tool, "denied", false)

		return ToolDecisionResponse{
			ToolCallID: req.ToolCallID,
			Status:     "denied",
			Message:    "Tool request denied. No changes were made.",
			Executed:   false,
			Tool:       tool,
			Preview:    buildToolActionPreview(tool, req.Input),
		}, nil
	}

	mode := echoapp.ParseMode(req.Mode)
	req.Input = echoapp.NormalizeToolInput(tool.Name, req.Input)
	if !echoapp.ToolAllowedForMode(mode, tool) {
		h.logDecision(ctx, req, tool, "mode_rejected", false)
		return ToolDecisionResponse{
			ToolCallID: req.ToolCallID,
			Status:     "error",
			Message:    "This mode cannot run that action. Switch to Agent mode for changes.",
			Executed:   false,
			Tool:       tool,
			Preview:    buildToolActionPreview(tool, req.Input),
		}, nil
	}
	if err := echoapp.ValidateToolInvocation(mode, tool, req.Input); err != nil {
		h.logDecision(ctx, req, tool, "invalid_input", false)
		return ToolDecisionResponse{
			ToolCallID: req.ToolCallID,
			Status:     "error",
			Message:    "The requested tool inputs were invalid. No changes were made.",
			Executed:   false,
			Tool:       tool,
			Preview:    buildToolActionPreview(tool, req.Input),
		}, nil
	}
	if h.RequirePendingApproval {
		if err := echoapp.ConsumePendingToolApproval(ctx, req.ToolCallID, tool, req.Input, mode); err != nil {
			h.logDecision(ctx, req, tool, "approval_rejected", false)
			return ToolDecisionResponse{
				ToolCallID: req.ToolCallID,
				Status:     "error",
				Message:    "This tool proposal is no longer valid or was already used. No changes were made.",
				Executed:   false,
				Tool:       tool,
				Preview:    buildToolActionPreview(tool, req.Input),
			}, nil
		}
	}

	// nosemgrep: trailofbits.go.invalid-usage-of-modified-variable.invalid-usage-of-modified-variable
	output, executed, duration, err := echoapp.ExecuteToolWithTimeout(ctx, h.executorChain(), tool.Name, req.Input)
	if err != nil {
		h.logDecision(ctx, req, tool, "execution_failed", false)
		return ToolDecisionResponse{
			ToolCallID: req.ToolCallID,
			Status:     "error",
			Message:    "Tool execution failed. No confirmed changes were made.",
			Executed:   false,
			Tool:       tool,
			Preview:    buildToolActionPreview(tool, req.Input),
			DurationMs: duration.Milliseconds(),
			Output: map[string]any{
				"error": userFacingToolError(err),
			},
		}, nil
	}
	if executed {
		h.logDecision(ctx, req, tool, "executed", true)
		message := "Tool executed successfully."
		if outputMessage, ok := stringValue(output["message"]); ok && outputMessage != "" {
			message = outputMessage
		}
		return ToolDecisionResponse{
			ToolCallID: req.ToolCallID,
			Status:     "success",
			Message:    message,
			Executed:   true,
			Tool:       tool,
			Preview:    buildToolActionPreview(tool, req.Input),
			Output:     output,
			DurationMs: duration.Milliseconds(),
		}, nil
	}

	h.logDecision(ctx, req, tool, "approved_unexecuted", false)
	message := "Approved, but this capability is not available in the current server configuration. No changes were made."
	if tool.Scope == "write" {
		message = "Approved, but this write capability has no server executor yet. No changes were made."
	}
	return ToolDecisionResponse{
		ToolCallID: req.ToolCallID,
		Status:     "error",
		Message:    message,
		Executed:   false,
		Tool:       tool,
		Preview:    buildToolActionPreview(tool, req.Input),
	}, nil
}

func (h *ToolDecisionHandler) executorChain() []ToolExecutor {
	out := make([]ToolExecutor, 0, len(h.Executors)+1)
	if h.Executor != nil {
		out = append(out, h.Executor)
	}
	out = append(out, h.Executors...)
	return out
}

func userFacingToolError(err error) string {
	if err == nil {
		return ""
	}
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return "unknown error"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		// Say so plainly: the model can retry with a narrower request or
		// another route, which it can't do for a cancelled one.
		return "the tool took too long and was stopped"
	}
	if errors.Is(err, context.Canceled) {
		return "request was interrupted"
	}
	return message
}

func (h *ToolDecisionHandler) logDecision(ctx context.Context, req *ToolDecisionRequest, tool ToolDef, status string, executed bool) {
	fields := []logger.Field{
		logger.String("tool_call_id", req.ToolCallID),
		logger.String("tool_name", tool.Name),
		logger.String("tool_scope", tool.Scope),
		logger.String("decision", strings.TrimSpace(strings.ToLower(req.Decision))),
		logger.String("status", status),
		logger.Bool("executed", executed),
	}
	device := authPkg.DeviceCtx(ctx)
	user, ok := authPkg.GetAuthUser(ctx)
	if (!ok || user == nil) && device.User != nil {
		user = device.User
		ok = true
	}
	if ok && user != nil {
		fields = append(fields,
			logger.Any("user_id", user.UserID),
			logger.Any("business_id", user.BusinessID),
			logger.Bool("is_console_admin", user.IsConsoleAdmin),
			logger.Bool("is_business_owner", user.EffectiveAccess != nil && user.EffectiveAccess.IsOwner),
			logger.Bool("is_business_admin", user.EffectiveAccess != nil && user.EffectiveAccess.IsAdmin),
		)
	}

	logger.Ctx(ctx).Info("caret tool decision recorded", fields...)
}

func toolDecisionValidationError(message string) error {
	return onekiterrors.New(http.StatusBadRequest, message)
}
