package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/application"
	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
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

func hasBusinessIdentity(ctx context.Context) bool {
	user, ok := authPkg.GetAuthUser(ctx)
	return ok && user != nil && user.UserID > 0 && user.BusinessID > 0
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
		ProjectId: sessionMetadataString(session.Metadata, "project_id"),
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
		Id:              strconv.Itoa(message.ID),
		Role:            message.Role,
		Content:         message.Content,
		CreatedAt:       message.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		AttachmentsJson: sessionImagesJSON(message.Metadata),
		TokensUsed:      int32(message.TokensUsed), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
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

	chatMessages, err := a.Service.SessionMessages(ctx, req.GetSessionId(), int(user.UserID))
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get chat messages")
	}

	messages := make([]*hubmemoryv1.ChatMessage, 0, len(chatMessages))
	for _, message := range chatMessages {
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

func (a *MemoryAPI) ListProjects(ctx context.Context, _ *hubmemoryv1.ListProjectsRequest) (*hubmemoryv1.ListProjectsResponse, error) {
	if !hasBusinessIdentity(ctx) {
		return nil, onekiterrors.PermissionDenied()
	}
	projects, err := a.Service.ListProjects(ctx)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list projects")
	}
	response := &hubmemoryv1.ListProjectsResponse{Projects: make([]*hubmemoryv1.CaretProject, 0, len(projects))}
	for _, project := range projects {
		response.Projects = append(response.Projects, projectToProto(project))
	}
	return response, nil
}

func (a *MemoryAPI) SaveProject(ctx context.Context, req *hubmemoryv1.SaveProjectRequest) (*hubmemoryv1.SaveProjectResponse, error) {
	if !hasBusinessIdentity(ctx) {
		return nil, onekiterrors.PermissionDenied()
	}
	project, err := a.Service.SaveProject(ctx, domain.Project{ID: req.GetId(), Name: req.GetName(), Icon: req.GetIcon(), Color: req.GetColor(), Context: req.GetContext()})
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid project")
	}
	return &hubmemoryv1.SaveProjectResponse{Project: projectToProto(*project)}, nil
}

func (a *MemoryAPI) ArchiveProject(ctx context.Context, req *hubmemoryv1.ArchiveProjectRequest) (*hubmemoryv1.ArchiveProjectResponse, error) {
	if !hasBusinessIdentity(ctx) {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.Service.ArchiveProject(ctx, req.GetId()); err != nil {
		if errors.Is(err, application.ErrChatSessionNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Project not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to archive project")
	}
	return &hubmemoryv1.ArchiveProjectResponse{Archived: true}, nil
}

func (a *MemoryAPI) SetChatProject(ctx context.Context, req *hubmemoryv1.SetChatProjectRequest) (*hubmemoryv1.SetChatProjectResponse, error) {
	user, ok := authPkg.GetAuthUser(ctx)
	if !ok || user == nil || user.UserID <= 0 || user.BusinessID <= 0 {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.Service.SetChatProject(ctx, req.GetSessionId(), int(user.UserID), req.GetProjectId()); err != nil {
		if errors.Is(err, application.ErrChatSessionNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Chat or project not found")
		}
		return nil, onekiterrors.New(http.StatusBadRequest, "Could not assign chat to project")
	}
	return &hubmemoryv1.SetChatProjectResponse{Updated: true}, nil
}

func projectToProto(project domain.Project) *hubmemoryv1.CaretProject {
	return &hubmemoryv1.CaretProject{Id: project.ID, Name: project.Name, Icon: project.Icon, Color: project.Color, Context: project.Context, UpdatedAt: project.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")}
}

func sessionMetadataString(raw, key string) string {
	var metadata map[string]json.RawMessage
	if json.Unmarshal([]byte(raw), &metadata) != nil {
		return ""
	}
	var value string
	_ = json.Unmarshal(metadata[key], &value)
	return value
}
func sessionImagesJSON(raw string) string {
	var metadata map[string]json.RawMessage
	if json.Unmarshal([]byte(raw), &metadata) != nil {
		return ""
	}
	return string(metadata["images"])
}
