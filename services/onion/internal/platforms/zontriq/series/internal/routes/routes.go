package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	seriesv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/series/v1"
	http "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platforms/zontriq/series/internal/application"
	"github.com/voxire/onion/internal/platforms/zontriq/series/internal/routes/handlers"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, seriesService *application.Service) {
	http.RegisterGET[seriesv1.ListSeriesRequest, seriesv1.ListSeriesResponse](
		ctx,
		router,
		handlers.ListSeriesRoutes(),
		&handlers.ListSeriesHandler{SeriesService: seriesService},
	)

	http.RegisterGET[seriesv1.GetSeriesRequest, seriesv1.GetSeriesResponse](
		ctx,
		router,
		handlers.GetSeriesRoutes(),
		&handlers.GetSeriesHandler{SeriesService: seriesService},
	)

	http.RegisterPOST[seriesv1.CreateSeriesRequest, seriesv1.CreateSeriesResponse](
		ctx,
		router,
		handlers.CreateSeriesRoutes(),
		&handlers.CreateSeriesHandler{SeriesService: seriesService},
	)

	http.RegisterPOST[seriesv1.UpdateSeriesRequest, seriesv1.UpdateSeriesResponse](
		ctx,
		router,
		handlers.UpdateSeriesRoutes(),
		&handlers.UpdateSeriesHandler{SeriesService: seriesService},
	)

	http.RegisterDELETE[seriesv1.DeleteSeriesRequest, seriesv1.DeleteSeriesResponse](
		ctx,
		router,
		handlers.DeleteSeriesRoutes(),
		&handlers.DeleteSeriesHandler{SeriesService: seriesService},
	)

	http.RegisterPOST[seriesv1.UpdateSeriesContextRequest, seriesv1.UpdateSeriesContextResponse](
		ctx,
		router,
		handlers.UpdateSeriesContextRoutes(),
		&handlers.UpdateSeriesContextHandler{SeriesService: seriesService},
	)
}
