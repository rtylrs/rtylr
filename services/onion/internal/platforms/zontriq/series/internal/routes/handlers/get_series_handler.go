package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	seriesv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/series/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/series/internal/application"
)

func GetSeriesRoutes() []string {
	return []string{
		"/zontriq/series/{id}",
	}
}

type GetSeriesHandler struct {
	SeriesService *application.Service
}

func (h *GetSeriesHandler) Serve(ctx context.Context, _ *seriesv1.GetSeriesRequest) (seriesv1.GetSeriesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return seriesv1.GetSeriesResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	idStr := ""
	if rctx := chi.RouteContext(ctx); rctx != nil {
		idStr = rctx.URLParam("id")
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return seriesv1.GetSeriesResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid series ID",
		}
	}

	series, seriesCtx, err := h.SeriesService.GetSeries(ctx, int32(id))
	if err != nil {
		return seriesv1.GetSeriesResponse{}, &errorModels.ServerError{
			Code:    http.StatusNotFound,
			Message: "Series not found",
		}
	}

	return seriesv1.GetSeriesResponse{
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
		Context: &seriesv1.SeriesContext{
			Id:                seriesCtx.Id,
			SeriesId:          seriesCtx.SeriesId,
			Context:           seriesCtx.Context,
			TopicsCovered:     seriesCtx.TopicsCovered,
			LastEpisodeNumber: seriesCtx.LastEpisodeNumber,
		},
	}, nil
}
