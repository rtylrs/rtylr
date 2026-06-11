package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/memory/internal/application"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func PinChatRoutes() []string { return []string{"/memory/chats/{sessionID}/pin"} }

type PinChatHandler struct{ Service *application.Service }

type PinChatRequest struct {
	Pinned bool `json:"pinned"`
}

type PinChatResponse struct {
	Pinned bool `json:"pinned"`
}

func (h *PinChatHandler) Serve(ctx context.Context, req *PinChatRequest) (PinChatResponse, *errorModels.ServerError) {
	user, ok := authenticatedUser(ctx)
	if !ok {
		return PinChatResponse{}, errorModels.PermissionDeniedServerError()
	}
	sessionID := chi.URLParamFromCtx(ctx, "sessionID")
	if sessionID == "" {
		return PinChatResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Missing sessionID",
		}
	}
	if _, err := h.Service.SetSessionPinned(ctx, sessionID, int(user.UserID), req.Pinned); err != nil {
		if errors.Is(err, application.ErrChatSessionNotFound) {
			return PinChatResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusNotFound,
				Code:     http.StatusNotFound,
				Message:  "Chat session not found",
			}
		}
		return PinChatResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to update chat session",
		}
	}
	return PinChatResponse{Pinned: req.Pinned}, nil
}
