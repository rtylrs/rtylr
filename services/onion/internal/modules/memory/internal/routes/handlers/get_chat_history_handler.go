package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/memory/internal/application"
	"github.com/voxire/onion/internal/modules/memory/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
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
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type ChatsResponse struct {
	Success  bool                 `json:"success"`
	UserID   int                  `json:"user_id,omitempty"`
	Sessions []domain.ChatSession `json:"sessions"`
	Limit    int                  `json:"limit"`
	Offset   int                  `json:"offset"`
	Error    string               `json:"error,omitempty"`
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

func (h *ChatHistoryHandler) Serve(ctx context.Context, req *ChatsRequest) (ChatsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ChatsResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, Message: "unauthorized"}
	}
	req.UserID = int(device.User.UserID)

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	sessions, err := h.MemoryService.ListUserSessions(ctx, req.UserID, limit, offset)
	if err != nil {
		return ChatsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list chat sessions",
		}
	}

	return ChatsResponse{
		Success:  true,
		UserID:   req.UserID,
		Sessions: sessions,
		Limit:    limit,
		Offset:   offset,
	}, nil
}

type ChatMessagesHandler struct {
	MemoryService *application.Service
}

func (h *ChatMessagesHandler) Serve(ctx context.Context, req *ChatMessagesRequest) (ChatMessagesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ChatMessagesResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, Message: "unauthorized"}
	}
	req.UserID = int(device.User.UserID)

	if req.SessionID == "" {
		return ChatMessagesResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Session ID is required",
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
