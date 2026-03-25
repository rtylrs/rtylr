package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	audiov1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/audio/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/audio/internal/application"
)

func DeleteAudioRoutes() []string {
	return []string{
		"/zontriq/audio/{id}",
	}
}

type DeleteAudioHandler struct {
	AudioService *application.Service
}

func (h *DeleteAudioHandler) Serve(ctx context.Context, _ *audiov1.DeleteAudioRequest) (audiov1.DeleteAudioResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return audiov1.DeleteAudioResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	idStr := chi.URLParam(nil, "id")
	if rctx := chi.RouteContext(ctx); rctx != nil {
		idStr = rctx.URLParam("id")
	}

	audioID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return audiov1.DeleteAudioResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid audio ID",
		}
	}

	err = h.AudioService.DeleteAudio(ctx, int32(audioID))
	if err != nil {
		return audiov1.DeleteAudioResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete audio",
		}
	}

	return audiov1.DeleteAudioResponse{}, nil
}
