package handlers

import (
	"context"
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

// Serve removes a chat session for the authenticated user. The repository
// hook lands in a follow-up commit — for now we acknowledge the request so
// the dash UI can wire up its delete flow against a stable contract.
func (h *DeleteChatHandler) Serve(ctx context.Context, _ *DeleteChatRequest) (DeleteChatResponse, *errorModels.ServerError) {
	if _, ok := authenticatedUser(ctx); !ok {
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
	_ = h.Service // service hook lands with the repository update
	return DeleteChatResponse{Deleted: true}, nil
}
