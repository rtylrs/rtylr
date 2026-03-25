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

func UpdateSeriesContextRoutes() []string {
	return []string{
		"/zontriq/series/{id}/context",
	}
}

type UpdateSeriesContextHandler struct {
	SeriesService *application.Service
}

func (h *UpdateSeriesContextHandler) Serve(ctx context.Context, req *seriesv1.UpdateSeriesContextRequest) (seriesv1.UpdateSeriesContextResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return seriesv1.UpdateSeriesContextResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	idStr := ""
	if rctx := chi.RouteContext(ctx); rctx != nil {
		idStr = rctx.URLParam("id")
	}

	seriesID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return seriesv1.UpdateSeriesContextResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid series ID",
		}
	}

	updatedCtx, updateErr := h.SeriesService.UpdateSeriesContext(ctx, int32(seriesID), req.GetContext())
	if updateErr != nil {
		return seriesv1.UpdateSeriesContextResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update series context",
		}
	}

	return seriesv1.UpdateSeriesContextResponse{
		Context: &seriesv1.SeriesContext{
			Id:                updatedCtx.Id,
			SeriesId:          updatedCtx.SeriesId,
			Context:           updatedCtx.Context,
			TopicsCovered:     updatedCtx.TopicsCovered,
			LastEpisodeNumber: updatedCtx.LastEpisodeNumber,
		},
	}, nil
}
