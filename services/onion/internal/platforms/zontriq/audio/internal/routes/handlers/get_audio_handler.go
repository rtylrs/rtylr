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

func GetAudioRoutes() []string {
	return []string{
		"/zontriq/audio/{id}",
	}
}

type GetAudioHandler struct {
	AudioService *application.Service
}

func (h *GetAudioHandler) Serve(ctx context.Context, _ *audiov1.GetAudioRequest) (audiov1.GetAudioResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return audiov1.GetAudioResponse{}, &errorModels.ServerError{
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
		return audiov1.GetAudioResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid audio ID",
		}
	}

	clip, err := h.AudioService.GetAudioByID(ctx, int32(audioID))
	if err != nil {
		return audiov1.GetAudioResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get audio",
		}
	}

	return audiov1.GetAudioResponse{
		Audio: &audiov1.AudioClip{
			Id:              int64(clip.ID),
			Title:           clip.Title,
			Script:          clip.Script,
			VoiceId:         clip.VoiceID,
			VoiceName:       clip.VoiceName,
			Provider:        clip.Provider,
			AudioUrl:        clip.AudioUrl,
			DurationSeconds: clip.DurationSeconds,
			State:           clip.State,
			ErrorMessage:    clip.ErrorMessage,
			CreatedAt:       clip.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:       clip.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
	}, nil
}
