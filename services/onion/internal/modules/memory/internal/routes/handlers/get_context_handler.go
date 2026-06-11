package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/memory/internal/application"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func ContextRoutes() []string {
	return []string{
		"/memory/context",
	}
}

type ContextRequest struct {
	SessionID string `json:"session_id"`
	UserID    int    `json:"user_id"`
}

type ContextResponse struct {
	Success bool        `json:"success"`
	Context interface{} `json:"context,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type ContextHandler struct {
	MemoryService *application.Service
}

func (h *ContextHandler) Serve(ctx context.Context, req *ContextRequest) (ContextResponse, *errorModels.ServerError) {
	user, ok := authenticatedUser(ctx)
	if !ok {
		return ContextResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, Message: "unauthorized"}
	}
	req.UserID = int(user.UserID)

	if req.SessionID == "" {
		return ContextResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Session ID is required",
		}
	}

	conversationContext, err := h.MemoryService.GetConversationContext(ctx, req.SessionID, req.UserID)
	if err != nil {
		return ContextResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get context",
		}
	}

	return ContextResponse{
		Success: true,
		Context: conversationContext,
	}, nil
}
