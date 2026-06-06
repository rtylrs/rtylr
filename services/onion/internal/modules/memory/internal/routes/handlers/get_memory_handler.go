package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/memory/internal/application"
	"github.com/voxire/onion/internal/modules/memory/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func MemoryRoutes() []string {
	return []string{
		"/memory/store",
	}
}

type MemoryRequest struct {
	UserID     int    `json:"user_id"`
	SessionID  string `json:"session_id"`
	Role       string `json:"role"`
	Content    string `json:"content"`
	TokensUsed int    `json:"tokens_used"`
}

type MemoryResponse struct {
	Success bool                        `json:"success"`
	Context *domain.ConversationContext `json:"context,omitempty"`
	Error   string                      `json:"error,omitempty"`
}

type MemoryHandler struct {
	MemoryService *application.Service
}

func (h *MemoryHandler) Validate(_ context.Context, req *MemoryRequest) *errorModels.ServerError {
	if req.SessionID == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Session ID is required",
		}
	}

	if req.Role == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Role is required",
		}
	}

	if req.Content == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Content is required",
		}
	}

	return nil
}

func (h *MemoryHandler) Serve(ctx context.Context, req *MemoryRequest) (MemoryResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return MemoryResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, Message: "unauthorized"}
	}
	req.UserID = int(device.User.UserID)

	_, err := h.MemoryService.StoreMessage(ctx, req.SessionID, req.UserID, req.Role, req.Content, req.TokensUsed)
	if err != nil {
		return MemoryResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to store message",
		}
	}

	conversationContext, err := h.MemoryService.GetConversationContext(ctx, req.SessionID, req.UserID)
	if err != nil {
		return MemoryResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get conversation context",
		}
	}

	return MemoryResponse{
		Success: true,
		Context: conversationContext,
	}, nil
}
