package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	audiov1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/audio/v1"
	http "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platforms/zontriq/audio/internal/application"
	"github.com/voxire/onion/internal/platforms/zontriq/audio/internal/routes/handlers"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, audioService *application.Service) {
	http.RegisterGET[audiov1.ListAudioRequest, audiov1.ListAudioResponse](
		ctx,
		router,
		handlers.ListAudioRoutes(),
		&handlers.ListAudioHandler{AudioService: audioService},
	)

	http.RegisterGET[audiov1.GetAudioRequest, audiov1.GetAudioResponse](
		ctx,
		router,
		handlers.GetAudioRoutes(),
		&handlers.GetAudioHandler{AudioService: audioService},
	)

	http.RegisterPOSTWithValidation[audiov1.GenerateAudioRequest, audiov1.GenerateAudioResponse](
		ctx,
		router,
		handlers.CreateAudioRoutes(),
		&handlers.CreateAudioHandler{AudioService: audioService},
	)

	http.RegisterDELETE[audiov1.DeleteAudioRequest, audiov1.DeleteAudioResponse](
		ctx,
		router,
		handlers.DeleteAudioRoutes(),
		&handlers.DeleteAudioHandler{AudioService: audioService},
	)
}
