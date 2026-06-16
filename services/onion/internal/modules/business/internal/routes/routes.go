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

	serverHttp.RegisterGETWithValidation[handlers.PublicBusinessBySlugRequest, handlers.PublicBusinessBySlugResponse](
		ctx, router, handlers.PublicBusinessBySlugRoutes(),
		&handlers.PublicBusinessBySlugHandler{Service: service},
	)

	serverHttp.RegisterPUTWithValidation[handlers.SetSlugRequest, handlers.SetSlugResponse](
		ctx, router, handlers.BusinessSlugRoutes(),
		&handlers.SetSlugHandler{Service: service},
	)

	serverHttp.RegisterGET[handlers.ListMyBusinessesRequest, handlers.ListMyBusinessesResponse](
		ctx, router, handlers.ListMyBusinessesRoutes(),
		&handlers.ListMyBusinessesHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.CreateSubBusinessRequest, handlers.CreateSubBusinessResponse](
		ctx, router, handlers.CreateSubBusinessRoutes(),
		&handlers.CreateSubBusinessHandler{Service: service},
	)
}
