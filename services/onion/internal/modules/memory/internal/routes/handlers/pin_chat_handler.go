package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/memory/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
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
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
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
	_ = h.Service
	return PinChatResponse{Pinned: req.Pinned}, nil
}
