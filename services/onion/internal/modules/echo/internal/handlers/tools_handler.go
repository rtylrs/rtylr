package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
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

type ToolExecutor interface {
	ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (output map[string]any, executed bool, err error)
}

type ToolDecisionHandler struct {
	Executor  ToolExecutor
	Executors []ToolExecutor
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
}

type ToolDecisionResponse struct {
	ToolCallID string         `json:"tool_call_id"`
	Status     string         `json:"status"`
	Message    string         `json:"message"`
	Executed   bool           `json:"executed"`
	Tool       ToolDef        `json:"tool"`
	Preview    string         `json:"preview,omitempty"`
	Output     map[string]any `json:"output,omitempty"`
}

func (h *ListToolsHandler) Serve(_ context.Context, req *ListToolsRequest) (ListToolsResponse, *errorModels.ServerError) {
	mode := ParseMode(req.Mode)
	return ListToolsResponse{Tools: ToolsForMode(mode)}, nil
}

func (h *ToolDecisionHandler) Validate(_ context.Context, req *ToolDecisionRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.ToolCallID) == "" {
		return toolDecisionValidationError("tool_call_id is required")
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

func (h *ToolDecisionHandler) Serve(ctx context.Context, req *ToolDecisionRequest) (ToolDecisionResponse, *errorModels.ServerError) {
	decision := strings.TrimSpace(strings.ToLower(req.Decision))
	name := strings.TrimSpace(req.Name)
	tool, ok := ToolByName(name)
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

	if decision == "deny" {
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

	if tool.Scope == "read" {
		h.logDecision(ctx, req, tool, "approved_read_unexecuted", false)
		return ToolDecisionResponse{
			ToolCallID: req.ToolCallID,
			Status:     "error",
			Message:    "Approved, but read-only tool execution from approval decisions is not available yet. Ask Caret to rerun the question so the server can ground it safely.",
			Executed:   false,
			Tool:       tool,
			Preview:    buildToolActionPreview(tool, req.Input),
		}, nil
	}

	for _, executor := range h.executorChain() {
		output, executed, err := executor.ExecuteCaretTool(ctx, tool.Name, req.Input)
		if err != nil {
			h.logDecision(ctx, req, tool, "execution_failed", false)
			return ToolDecisionResponse{
				ToolCallID: req.ToolCallID,
				Status:     "error",
				Message:    "Tool execution failed. No confirmed changes were made.",
				Executed:   false,
				Tool:       tool,
				Preview:    buildToolActionPreview(tool, req.Input),
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
			}, nil
		}
	}

	h.logDecision(ctx, req, tool, "approved_unexecuted", false)
	return ToolDecisionResponse{
		ToolCallID: req.ToolCallID,
		Status:     "error",
		Message:    "Approved, but this write tool does not have a server executor yet. No changes were made.",
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
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
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
	if user, ok := authPkg.GetAuthUser(ctx); ok && user != nil {
		fields = append(fields,
			logger.Any("user_id", user.UserID),
			logger.Any("business_id", user.BusinessID),
			logger.String("role", user.Role),
		)
	}

	logger.Ctx(ctx).Info("caret tool decision recorded", fields...)
}

func toolDecisionValidationError(message string) *errorModels.ServerError {
	return &errorModels.ServerError{
		HTTPCode: http.StatusBadRequest,
		Code:     http.StatusBadRequest,
		Message:  message,
	}
}
