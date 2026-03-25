package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	voicesv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/voices/v1"
	http "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platforms/zontriq/voices/internal/routes/handlers"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux) {
	http.RegisterGET[voicesv1.ListVoicesRequest, voicesv1.ListVoicesResponse](
		ctx,
		router,
		handlers.ListVoicesRoutes(),
		&handlers.ListVoicesHandler{},
	)
}
