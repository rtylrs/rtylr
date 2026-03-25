package handlers

import (
	"context"
	"net/http"

	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platforms/rtylr/memory/internal/application"
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
	if req.SessionID == "" {
		return SummarizeResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Session ID is required",
		}
	}

	if req.UserID <= 0 {
		return SummarizeResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "User ID is required",
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
