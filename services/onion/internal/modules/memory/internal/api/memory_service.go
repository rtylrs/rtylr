package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/modules/memory/internal/application"
	"github.com/voxire/onion/internal/modules/memory/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hubmemoryv1 "github.com/voxire/voxie/gen/go/hub/memory/v1"
)

type MemoryAPI struct {
	Service *application.Service
}

func NewMemoryAPI(service *application.Service) *MemoryAPI {
	return &MemoryAPI{Service: service}
}

func sessionToChatSummary(session domain.ChatSession) *hubmemoryv1.ChatSummary {
	title := session.Title
	if title == "" {
		title = "Untitled chat"
	}
	return &hubmemoryv1.ChatSummary{
		Id:        session.SessionID,
		Title:     title,
		Summary:   session.Summary,
		UpdatedAt: session.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Pinned:    sessionMetadataBool(session.Metadata, "pinned"),
		Archived:  sessionMetadataBool(session.Metadata, "archived"),
	}
}

func sessionMetadataBool(raw, key string) bool {
	if raw == "" || raw == "null" {
		return false
	}
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &metadata); err != nil {
		return false
	}
	value, ok := metadata[key].(bool)
	return ok && value
}

func messageToProto(message domain.ConversationMessage) *hubmemoryv1.ChatMessage {
	return &hubmemoryv1.ChatMessage{
		Id:         strconv.Itoa(message.ID),
		Role:       message.Role,
		Content:    message.Content,
		CreatedAt:  message.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		TokensUsed: int32(message.TokensUsed), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	}
}

func (a *MemoryAPI) ListChats(ctx context.Context, req *hubmemoryv1.ListChatsRequest) (*hubmemoryv1.ListChatsResponse, error) {
	user, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := int(req.GetOffset())
	if offset < 0 {
		offset = 0
	}

	sessions, err := a.Service.ListUserSessions(ctx, int(user.UserID), limit, offset)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list chat sessions")
	}

	chats := make([]*hubmemoryv1.ChatSummary, 0, len(sessions))
	for _, session := range sessions {
		chats = append(chats, sessionToChatSummary(session))
	}
	return &hubmemoryv1.ListChatsResponse{Chats: chats}, nil
}

func (a *MemoryAPI) GetChatMessages(ctx context.Context, req *hubmemoryv1.GetChatMessagesRequest) (*hubmemoryv1.GetChatMessagesResponse, error) {
	user, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	if req.GetSessionId() == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Session ID is required")
	}

	conversationContext, err := a.Service.GetConversationContext(ctx, req.GetSessionId(), int(user.UserID))
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get chat messages")
	}

	messages := make([]*hubmemoryv1.ChatMessage, 0, len(conversationContext.Messages))
	for _, message := range conversationContext.Messages {
		messages = append(messages, messageToProto(message))
	}
	return &hubmemoryv1.GetChatMessagesResponse{Messages: messages}, nil
}

func (a *MemoryAPI) DeleteChat(ctx context.Context, req *hubmemoryv1.DeleteChatRequest) (*hubmemoryv1.DeleteChatResponse, error) {
	user, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.Service.DeleteSession(ctx, req.GetSessionId(), int(user.UserID)); err != nil {
		if errors.Is(err, application.ErrChatSessionNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Chat session not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete chat session")
	}
	return &hubmemoryv1.DeleteChatResponse{Deleted: true}, nil
}

func (a *MemoryAPI) RenameChat(ctx context.Context, req *hubmemoryv1.RenameChatRequest) (*hubmemoryv1.RenameChatResponse, error) {
	user, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	if req.GetTitle() == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Title is required")
	}
	if len(req.GetTitle()) > 120 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Title is too long")
	}
	if _, err := a.Service.RenameSession(ctx, req.GetSessionId(), int(user.UserID), req.GetTitle()); err != nil {
		if errors.Is(err, application.ErrChatSessionNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Chat session not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to rename chat session")
	}
	return &hubmemoryv1.RenameChatResponse{Updated: true}, nil
}

func (a *MemoryAPI) PinChat(ctx context.Context, req *hubmemoryv1.PinChatRequest) (*hubmemoryv1.PinChatResponse, error) {
	user, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	if _, err := a.Service.SetSessionPinned(ctx, req.GetSessionId(), int(user.UserID), req.GetPinned()); err != nil {
		if errors.Is(err, application.ErrChatSessionNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Chat session not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update chat session")
	}
	return &hubmemoryv1.PinChatResponse{Updated: true}, nil
}
