package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/memory/internal/application"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	authPkg "github.com/voxire/onion/internal/platform/auth"
)

func StatsRoutes() []string {
	return []string{
		"/memory/stats",
	}
}

type StatsRequest struct {
	UserID int `json:"user_id"`
}

type StatsResponse struct {
	Success bool        `json:"success"`
	Stats   interface{} `json:"stats,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type StatsHandler struct {
	MemoryService *application.Service
}

func (h *StatsHandler) Serve(ctx context.Context, req *StatsRequest) (StatsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return StatsResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, Message: "unauthorized"}
	}
	req.UserID = int(device.User.UserID)

	stats, err := h.MemoryService.GetUserStats(ctx, req.UserID)
	if err != nil {
		return StatsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get user stats",
		}
	}

	return StatsResponse{
		Success: true,
		Stats:   stats,
	}, nil
}
