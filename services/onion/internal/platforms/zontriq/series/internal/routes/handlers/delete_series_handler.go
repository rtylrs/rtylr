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

func DeleteSeriesRoutes() []string {
	return []string{
		"/zontriq/series/{id}",
	}
}

type DeleteSeriesHandler struct {
	SeriesService *application.Service
}

func (h *DeleteSeriesHandler) Serve(ctx context.Context, _ *seriesv1.DeleteSeriesRequest) (seriesv1.DeleteSeriesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return seriesv1.DeleteSeriesResponse{}, &errorModels.ServerError{
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
		return seriesv1.DeleteSeriesResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid series ID",
		}
	}

	deleteErr := h.SeriesService.DeleteSeries(ctx, int32(id))
	if deleteErr != nil {
		return seriesv1.DeleteSeriesResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete series",
		}
	}

	return seriesv1.DeleteSeriesResponse{}, nil
}
