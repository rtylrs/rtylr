package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	videosv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/videos/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/videos/internal/application"
)

func CreateVideoRoutes() []string {
	return []string{
		"/zontriq/videos",
	}
}

type CreateVideoHandler struct {
	VideoService *application.Service
}

func (h *CreateVideoHandler) Validate(_ context.Context, req *videosv1.CreateVideoRequest) *errorModels.ServerError {
	if req.GetTitle() == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Title is required",
		}
	}

	return nil
}

func (h *CreateVideoHandler) Serve(ctx context.Context, req *videosv1.CreateVideoRequest) (videosv1.CreateVideoResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return videosv1.CreateVideoResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	var seriesID *int32
	if req.GetSeriesId() > 0 {
		sid := int32(req.GetSeriesId())
		seriesID = &sid
	}

	var configJSON []byte

	if req.GetConfig() != nil {
		var marshalErr error

		configJSON, marshalErr = json.Marshal(req.GetConfig())
		if marshalErr != nil {
			configJSON = []byte("{}")
		}
	} else {
		configJSON = []byte("{}")
	}

	video, err := h.VideoService.CreateVideo(
		ctx,
		int32(device.User.PlatformUserID),
		seriesID,
		req.GetTitle(),
		req.GetScript(),
		configJSON,
		req.GetChannelIds(),
	)
	if err != nil {
		return videosv1.CreateVideoResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create video",
		}
	}

	return videosv1.CreateVideoResponse{
		Video: &videosv1.Video{
			Id:        int64(video.ID),
			SeriesId:  int64(video.SeriesID),
			Title:     video.Title,
			State:     video.State,
			CreatedAt: video.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: video.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
	}, nil
}
