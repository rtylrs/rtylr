package handlers

import (
	"context"
	"net/http"

	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/modules/memory/internal/application"
)

func ChatsRoutes() []string {
	return []string{
		"/memory/chats",
	}
}

func ChatMessagesRoutes() []string {
	return []string{
		"/memory/chats/{sessionID}/messages",
	}
}

type ChatsRequest struct {
	UserID int `json:"user_id"`
}

type ChatsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	UserID  int    `json:"user_id,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ChatMessagesRequest struct {
	SessionID string `json:"session_id"`
	UserID    int    `json:"user_id"`
}

type ChatMessagesResponse struct {
	Success   bool        `json:"success"`
	SessionID string      `json:"session_id,omitempty"`
	UserID    int         `json:"user_id,omitempty"`
	Context   interface{} `json:"context,omitempty"`
	Error     string      `json:"error,omitempty"`
}

type ChatHistoryHandler struct {
	MemoryService *application.Service
}

func (h *ChatHistoryHandler) Serve(_ context.Context, req *ChatsRequest) (ChatsResponse, *errorModels.ServerError) {
	if req.UserID <= 0 {
		return ChatsResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "User ID is required",
		}
	}

	// For now, return a simple response.
	// TODO: Implement ListUserSessions in the service.
	return ChatsResponse{
		Success: true,
		Message: "User chats functionality coming soon",
		UserID:  req.UserID,
	}, nil
}

type ChatMessagesHandler struct {
	MemoryService *application.Service
}

func (h *ChatMessagesHandler) Serve(ctx context.Context, req *ChatMessagesRequest) (ChatMessagesResponse, *errorModels.ServerError) {
	if req.SessionID == "" {
		return ChatMessagesResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Session ID is required",
		}
	}

	if req.UserID <= 0 {
		return ChatMessagesResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "User ID is required",
		}
	}

	conversationContext, err := h.MemoryService.GetConversationContext(ctx, req.SessionID, req.UserID)
	if err != nil {
		return ChatMessagesResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get chat messages",
		}
	}

	return ChatMessagesResponse{
		Success:   true,
		SessionID: req.SessionID,
		UserID:    req.UserID,
		Context:   conversationContext,
	}, nil
}
