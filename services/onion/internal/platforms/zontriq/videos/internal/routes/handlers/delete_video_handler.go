package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	videosv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/videos/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/videos/internal/application"
)

func DeleteVideoRoutes() []string {
	return []string{
		"/zontriq/videos/{id}",
	}
}

type DeleteVideoHandler struct {
	VideoService *application.Service
}

func (h *DeleteVideoHandler) Serve(ctx context.Context, _ *videosv1.DeleteVideoRequest) (videosv1.DeleteVideoResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return videosv1.DeleteVideoResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	idStr := chi.URLParam(nil, "id")
	if rctx := chi.RouteContext(ctx); rctx != nil {
		idStr = rctx.URLParam("id")
	}

	videoID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return videosv1.DeleteVideoResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid video ID",
		}
	}

	err = h.VideoService.DeleteVideo(ctx, int32(videoID))
	if err != nil {
		return videosv1.DeleteVideoResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete video",
		}
	}

	return videosv1.DeleteVideoResponse{}, nil
}
