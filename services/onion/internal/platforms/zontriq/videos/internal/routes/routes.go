package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	videosv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/videos/v1"
	http "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platforms/zontriq/videos/internal/application"
	"github.com/voxire/onion/internal/platforms/zontriq/videos/internal/routes/handlers"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, videoService *application.Service) {
	http.RegisterGET[videosv1.ListVideosRequest, videosv1.ListVideosResponse](
		ctx,
		router,
		handlers.ListVideosRoutes(),
		&handlers.ListVideosHandler{VideoService: videoService},
	)

	http.RegisterGET[videosv1.GetVideoRequest, videosv1.GetVideoResponse](
		ctx,
		router,
		handlers.GetVideoRoutes(),
		&handlers.GetVideoHandler{VideoService: videoService},
	)

	http.RegisterPOSTWithValidation[videosv1.CreateVideoRequest, videosv1.CreateVideoResponse](
		ctx,
		router,
		handlers.CreateVideoRoutes(),
		&handlers.CreateVideoHandler{VideoService: videoService},
	)

	http.RegisterDELETE[videosv1.DeleteVideoRequest, videosv1.DeleteVideoResponse](
		ctx,
		router,
		handlers.DeleteVideoRoutes(),
		&handlers.DeleteVideoHandler{VideoService: videoService},
	)
}
