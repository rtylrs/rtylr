package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/memory/internal/application"
	"github.com/voxire/onion/internal/modules/memory/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func authenticatedUser(ctx context.Context) (*authPkg.AuthUser, bool) {
	if user, ok := authPkg.GetAuthUser(ctx); ok && user != nil && user.UserID > 0 {
		return user, true
	}

	device := authPkg.DeviceCtx(ctx)
	if device.User != nil && device.User.UserID > 0 {
		return device.User, true
	}

	return nil, false
}

type SessionRequest struct {
	UserID int    `json:"user_id"`
	Title  string `json:"title"`
}

type SessionResponse struct {
	Success bool                `json:"success"`
	Session *domain.ChatSession `json:"session,omitempty"`
	Error   string              `json:"error,omitempty"`
}

type SessionHandler struct {
	MemoryService *application.Service
}

func (h *SessionHandler) Validate(_ context.Context, req *SessionRequest) error {
	if req.Title == "" {
		return onekiterrors.New(http.StatusBadRequest, "Session title is required")
	}
	return nil
}

func (h *SessionHandler) Serve(ctx context.Context, req *SessionRequest) (SessionResponse, error) {
	user, ok := authenticatedUser(ctx)
	if !ok {
		return SessionResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	req.UserID = int(user.UserID)

	session, err := h.MemoryService.CreateSession(ctx, req.UserID, req.Title)
	if err != nil {
		return SessionResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create session")
	}

	return SessionResponse{Success: true, Session: session}, nil
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

func (h *MemoryHandler) Validate(_ context.Context, req *MemoryRequest) error {
	if req.SessionID == "" {
		return onekiterrors.New(http.StatusBadRequest, "Session ID is required")
	}
	if req.Role == "" {
		return onekiterrors.New(http.StatusBadRequest, "Role is required")
	}
	if req.Content == "" {
		return onekiterrors.New(http.StatusBadRequest, "Content is required")
	}
	return nil
}

func (h *MemoryHandler) Serve(ctx context.Context, req *MemoryRequest) (MemoryResponse, error) {
	user, ok := authenticatedUser(ctx)
	if !ok {
		return MemoryResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	req.UserID = int(user.UserID)

	if _, err := h.MemoryService.StoreMessage(ctx, req.SessionID, req.UserID, req.Role, req.Content, req.TokensUsed); err != nil {
		return MemoryResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to store message")
	}

	conversationContext, err := h.MemoryService.GetConversationContext(ctx, req.SessionID, req.UserID)
	if err != nil {
		return MemoryResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get conversation context")
	}

	return MemoryResponse{Success: true, Context: conversationContext}, nil
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

func (h *ContextHandler) Serve(ctx context.Context, req *ContextRequest) (ContextResponse, error) {
	user, ok := authenticatedUser(ctx)
	if !ok {
		return ContextResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	req.UserID = int(user.UserID)

	if req.SessionID == "" {
		return ContextResponse{}, onekiterrors.New(http.StatusBadRequest, "Session ID is required")
	}

	conversationContext, err := h.MemoryService.GetConversationContext(ctx, req.SessionID, req.UserID)
	if err != nil {
		return ContextResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get context")
	}

	return ContextResponse{Success: true, Context: conversationContext}, nil
}

type SearchRequest struct {
	UserID int    `json:"user_id"`
	Query  string `json:"q"`
}

type SearchResponse struct {
	Success  bool        `json:"success"`
	Memories interface{} `json:"memories,omitempty"`
	Error    string      `json:"error,omitempty"`
}

type SearchHandler struct {
	MemoryService *application.Service
}

func (h *SearchHandler) Serve(ctx context.Context, req *SearchRequest) (SearchResponse, error) {
	user, ok := authenticatedUser(ctx)
	if !ok {
		return SearchResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	req.UserID = int(user.UserID)

	if req.Query == "" {
		return SearchResponse{}, onekiterrors.New(http.StatusBadRequest, "Search query is required")
	}

	memories, err := h.MemoryService.SearchMemories(ctx, req.UserID, req.Query)
	if err != nil {
		return SearchResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to search memories")
	}

	return SearchResponse{Success: true, Memories: memories}, nil
}

type StatsRequest struct {
	UserID int `json:"user_id"`
}

type StatsResponse struct {
	Success bool        `json:"success"`
	Stats   interface{} `json:"stats,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type StatsHandler struct {
	MemoryService *application.Service
}

func (h *StatsHandler) Serve(ctx context.Context, req *StatsRequest) (StatsResponse, error) {
	user, ok := authenticatedUser(ctx)
	if !ok {
		return StatsResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	req.UserID = int(user.UserID)

	stats, err := h.MemoryService.GetUserStats(ctx, req.UserID)
	if err != nil {
		return StatsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get user stats")
	}

	return StatsResponse{Success: true, Stats: stats}, nil
}

type SummarizeRequest struct {
	SessionID string `json:"session_id"`
	UserID    int    `json:"user_id"`
}

type SummarizeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type SummarizeHandler struct {
	MemoryService *application.Service
}

func (h *SummarizeHandler) Serve(ctx context.Context, req *SummarizeRequest) (SummarizeResponse, error) {
	user, ok := authenticatedUser(ctx)
	if !ok {
		return SummarizeResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	req.UserID = int(user.UserID)

	if req.SessionID == "" {
		return SummarizeResponse{}, onekiterrors.New(http.StatusBadRequest, "Session ID is required")
	}

	if err := h.MemoryService.SummarizeSession(ctx, req.SessionID, req.UserID); err != nil {
		return SummarizeResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to summarize session")
	}

	return SummarizeResponse{Success: true, Message: "Session summarized successfully"}, nil
}
