package handlers

import (
	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	seriesv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/series/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/series/internal/application"
)

func CreateSeriesRoutes() []string {
	return []string{
		"/zontriq/series",
	}
}

type CreateSeriesHandler struct {
	SeriesService *application.Service
}

func (h *CreateSeriesHandler) Serve(ctx context.Context, req *seriesv1.CreateSeriesRequest) (seriesv1.CreateSeriesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return seriesv1.CreateSeriesResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	userID := int32(device.User.PlatformUserID)

	series, err := h.SeriesService.CreateSeries(
		ctx,
		userID,
		req.GetName(),
		req.GetDescription(),
		req.GetNiche(),
		req.GetVoice(),
		req.GetStyle(),
		req.GetFrequency(),
		req.GetConfig(),
		req.GetPublishTime(),
	)
	if err != nil {
		return seriesv1.CreateSeriesResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create series",
		}
	}

	return seriesv1.CreateSeriesResponse{
		Series: &seriesv1.Series{
			Id:          series.Id,
			Name:        series.Name,
			Description: series.Description,
			Niche:       series.Niche,
			Voice:       series.Voice,
			Style:       series.Style,
			Frequency:   series.Frequency,
			Config:      series.Config,
			PublishTime: series.PublishTime,
			State:       series.State,
			VideoCount:  series.VideoCount,
			Status:      series.Status,
			CreatedAt:   series.CreatedAt,
			UpdatedAt:   series.UpdatedAt,
		},
	}, nil
}
