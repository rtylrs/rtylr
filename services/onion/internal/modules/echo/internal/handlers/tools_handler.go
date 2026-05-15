package handlers

import (
	"context"

	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ListToolsRoutes is the HTTP path for the tool registry endpoint.
func ListToolsRoutes() []string { return []string{"/caret/tools"} }

// ListToolsHandler returns the tools available to the calling user. Today the
// list is a static, mode-agnostic seed registry — the eventual plan is to
// gate tools by the caller's role + active plan.
type ListToolsHandler struct{}

type ListToolsRequest struct {
	Mode string `json:"mode"`
}

type ListToolsResponse struct {
	Tools []ToolDef `json:"tools"`
}

func (h *ListToolsHandler) Serve(_ context.Context, req *ListToolsRequest) (ListToolsResponse, *errorModels.ServerError) {
	mode := ParseMode(req.Mode)
	return ListToolsResponse{Tools: ToolsForMode(mode)}, nil
}
