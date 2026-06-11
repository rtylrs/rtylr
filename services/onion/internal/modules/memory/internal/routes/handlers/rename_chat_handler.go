package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/memory/internal/application"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func RenameChatRoutes() []string { return []string{"/memory/chats/{sessionID}"} }

type RenameChatHandler struct{ Service *application.Service }

type RenameChatRequest struct {
	Title string `json:"title"`
}

type RenameChatResponse struct {
	Updated bool `json:"updated"`
}

// Validate enforces a non-empty, length-capped title server-side.
func (h *RenameChatHandler) Validate(_ context.Context, req *RenameChatRequest) *errorModels.ServerError {
	if req.Title == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Title is required",
		}
	}
	if len(req.Title) > 120 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Title is too long",
		}
	}
	return nil
}

func (h *RenameChatHandler) Serve(ctx context.Context, req *RenameChatRequest) (RenameChatResponse, *errorModels.ServerError) {
	user, ok := authenticatedUser(ctx)
	if !ok {
		return RenameChatResponse{}, errorModels.PermissionDeniedServerError()
	}
	sessionID := chi.URLParamFromCtx(ctx, "sessionID")
	if sessionID == "" {
		return RenameChatResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Missing sessionID",
		}
	}
	if _, err := h.Service.RenameSession(ctx, sessionID, int(user.UserID), req.Title); err != nil {
		if errors.Is(err, application.ErrChatSessionNotFound) {
			return RenameChatResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusNotFound,
				Code:     http.StatusNotFound,
				Message:  "Chat session not found",
			}
		}
		return RenameChatResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to rename chat session",
		}
	}
	return RenameChatResponse{Updated: true}, nil
}
