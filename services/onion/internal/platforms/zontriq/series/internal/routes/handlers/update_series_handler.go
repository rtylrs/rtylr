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

func UpdateSeriesRoutes() []string {
	return []string{
		"/zontriq/series/{id}",
	}
}

type UpdateSeriesHandler struct {
	SeriesService *application.Service
}

func (h *UpdateSeriesHandler) Serve(ctx context.Context, req *seriesv1.UpdateSeriesRequest) (seriesv1.UpdateSeriesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return seriesv1.UpdateSeriesResponse{}, &errorModels.ServerError{
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
		return seriesv1.UpdateSeriesResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid series ID",
		}
	}

	series, updateErr := h.SeriesService.UpdateSeries(
		ctx,
		int32(id),
		req.GetName(),
		req.GetDescription(),
		req.GetNiche(),
		req.GetVoice(),
		req.GetStyle(),
		req.GetFrequency(),
		req.GetConfig(),
		req.GetPublishTime(),
		req.GetState(),
	)
	if updateErr != nil {
		return seriesv1.UpdateSeriesResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update series",
		}
	}

	return seriesv1.UpdateSeriesResponse{
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
