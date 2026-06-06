package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/memory/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func SummarizeRoutes() []string {
	return []string{
		"/memory/summarize",
	}
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

func (h *SummarizeHandler) Serve(ctx context.Context, req *SummarizeRequest) (SummarizeResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return SummarizeResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, Message: "unauthorized"}
	}
	req.UserID = int(device.User.UserID)

	if req.SessionID == "" {
		return SummarizeResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Session ID is required",
		}
	}

	err := h.MemoryService.SummarizeSession(ctx, req.SessionID, req.UserID)
	if err != nil {
		return SummarizeResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to summarize session",
		}
	}

	return SummarizeResponse{
		Success: true,
		Message: "Session summarized successfully",
	}, nil
}
