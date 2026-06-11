package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/memory/internal/application"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func DeleteChatRoutes() []string { return []string{"/memory/chats/{sessionID}"} }

type DeleteChatHandler struct{ Service *application.Service }

type DeleteChatRequest struct{}

type DeleteChatResponse struct {
	Deleted bool `json:"deleted"`
}

// Serve removes a chat session and its stored conversation rows for the
// authenticated user.
func (h *DeleteChatHandler) Serve(ctx context.Context, _ *DeleteChatRequest) (DeleteChatResponse, *errorModels.ServerError) {
	user, ok := authenticatedUser(ctx)
	if !ok {
		return DeleteChatResponse{}, errorModels.PermissionDeniedServerError()
	}
	sessionID := chi.URLParamFromCtx(ctx, "sessionID")
	if sessionID == "" {
		return DeleteChatResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Missing sessionID",
		}
	}
	if err := h.Service.DeleteSession(ctx, sessionID, int(user.UserID)); err != nil {
		if errors.Is(err, application.ErrChatSessionNotFound) {
			return DeleteChatResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusNotFound,
				Code:     http.StatusNotFound,
				Message:  "Chat session not found",
			}
		}
		return DeleteChatResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to delete chat session",
		}
	}
	return DeleteChatResponse{Deleted: true}, nil
}
