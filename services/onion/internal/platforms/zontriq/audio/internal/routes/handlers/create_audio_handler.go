package handlers

import (
	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	audiov1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/audio/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/audio/internal/application"
)

func CreateAudioRoutes() []string {
	return []string{
		"/zontriq/audio",
	}
}

type CreateAudioHandler struct {
	AudioService *application.Service
}

func (h *CreateAudioHandler) Validate(_ context.Context, req *audiov1.GenerateAudioRequest) *errorModels.ServerError {
	if req.GetTitle() == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Title is required",
		}
	}

	if req.GetScript() == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Script is required",
		}
	}

	if req.GetVoiceId() == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Voice ID is required",
		}
	}

	if req.GetProvider() == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Provider is required",
		}
	}

	return nil
}

func (h *CreateAudioHandler) Serve(ctx context.Context, req *audiov1.GenerateAudioRequest) (audiov1.GenerateAudioResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return audiov1.GenerateAudioResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	clip, err := h.AudioService.CreateAudioClip(
		ctx,
		int32(device.User.PlatformUserID),
		req.GetTitle(),
		req.GetScript(),
		req.GetVoiceId(),
		"",
		req.GetProvider(),
	)
	if err != nil {
		return audiov1.GenerateAudioResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create audio clip",
		}
	}

	return audiov1.GenerateAudioResponse{
		Audio: &audiov1.AudioClip{
			Id:        int64(clip.ID),
			Title:     clip.Title,
			Script:    clip.Script,
			VoiceId:   clip.VoiceID,
			VoiceName: clip.VoiceName,
			Provider:  clip.Provider,
			State:     clip.State,
			CreatedAt: clip.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: clip.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
	}, nil
}
