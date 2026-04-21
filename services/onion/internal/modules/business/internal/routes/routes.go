package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/business/internal/application"
	"github.com/voxire/onion/internal/modules/business/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.GetProfileRequest, handlers.GetProfileResponse](
		ctx, router, handlers.BusinessMeRoutes(),
		&handlers.GetProfileHandler{Service: service},
	)

	// Anonymous public read — registered in module.go's
	// GetAnonymousRoutes so the auth middleware skips it.
	serverHttp.RegisterGETWithValidation[handlers.PublicBusinessRequest, handlers.PublicBusinessResponse](
		ctx, router, handlers.PublicBusinessRoutes(),
		&handlers.PublicBusinessHandler{Service: service},
	)
}
