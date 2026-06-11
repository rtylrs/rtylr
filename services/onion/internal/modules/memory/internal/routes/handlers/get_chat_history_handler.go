package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/memory/internal/application"
	"github.com/voxire/onion/internal/modules/memory/internal/domain"
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
	Chats    []ChatSummary        `json:"chats"`
	Limit    int                  `json:"limit"`
	Offset   int                  `json:"offset"`
	Error    string               `json:"error,omitempty"`
}

type ChatSummary struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
	Pinned    bool      `json:"pinned,omitempty"`
	Archived  bool      `json:"archived,omitempty"`
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
	Messages  []Message   `json:"messages"`
	Error     string      `json:"error,omitempty"`
}

type Message struct {
	ID         string    `json:"id"`
	Role       string    `json:"role"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"createdAt"`
	TokensUsed int       `json:"tokensUsed,omitempty"`
}

type ChatHistoryHandler struct {
	MemoryService *application.Service
}

func (h *ChatHistoryHandler) Serve(ctx context.Context, req *ChatsRequest) (ChatsResponse, *errorModels.ServerError) {
	user, ok := authenticatedUser(ctx)
	if !ok {
		return ChatsResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, Message: "unauthorized"}
	}
	req.UserID = int(user.UserID)

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
		Chats:    toChatSummaries(sessions),
		Limit:    limit,
		Offset:   offset,
	}, nil
}

type ChatMessagesHandler struct {
	MemoryService *application.Service
}

func (h *ChatMessagesHandler) Serve(ctx context.Context, req *ChatMessagesRequest) (ChatMessagesResponse, *errorModels.ServerError) {
	user, ok := authenticatedUser(ctx)
	if !ok {
		return ChatMessagesResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, Message: "unauthorized"}
	}
	req.UserID = int(user.UserID)
	if pathSessionID := chi.URLParamFromCtx(ctx, "sessionID"); pathSessionID != "" {
		req.SessionID = pathSessionID
	}

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
		Messages:  toMessages(conversationContext.Messages),
	}, nil
}

func toChatSummaries(sessions []domain.ChatSession) []ChatSummary {
	chats := make([]ChatSummary, 0, len(sessions))
	for _, session := range sessions {
		title := session.Title
		if title == "" {
			title = "Untitled chat"
		}
		chats = append(chats, ChatSummary{
			ID:        session.SessionID,
			Title:     title,
			Summary:   session.Summary,
			UpdatedAt: session.UpdatedAt,
		})
	}
	return chats
}

func toMessages(messages []domain.ConversationMessage) []Message {
	out := make([]Message, 0, len(messages))
	for _, message := range messages {
		out = append(out, Message{
			ID:         strconv.Itoa(message.ID),
			Role:       message.Role,
			Content:    message.Content,
			CreatedAt:  message.CreatedAt,
			TokensUsed: message.TokensUsed,
		})
	}
	return out
}
