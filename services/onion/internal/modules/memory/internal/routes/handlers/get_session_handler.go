package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/memory/internal/application"
	"github.com/voxire/onion/internal/modules/memory/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func SessionRoutes() []string {
	return []string{
		"/memory/sessions",
	}
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

func (h *SessionHandler) Validate(_ context.Context, req *SessionRequest) *errorModels.ServerError {
	if req.Title == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Session title is required",
		}
	}

	return nil
}

func (h *SessionHandler) Serve(ctx context.Context, req *SessionRequest) (SessionResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return SessionResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, Message: "unauthorized"}
	}
	req.UserID = int(device.User.UserID)

	session, err := h.MemoryService.CreateSession(ctx, req.UserID, req.Title)
	if err != nil {
		return SessionResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create session",
		}
	}

	return SessionResponse{
		Success: true,
		Session: session,
	}, nil
}
