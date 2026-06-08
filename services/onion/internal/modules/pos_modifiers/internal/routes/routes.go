package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_modifiers/internal/application"
	"github.com/voxire/onion/internal/modules/pos_modifiers/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.ListModifiersRequest, handlers.ListModifiersResponse](
		ctx, router, handlers.ListModifiersRoutes(),
		&handlers.ListModifiersHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CreateModifierRequest, handlers.CreateModifierResponse](
		ctx, router, handlers.CreateModifierRoutes(),
		&handlers.CreateModifierHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[handlers.UpdateModifierRequest, handlers.UpdateModifierResponse](
		ctx, router, handlers.UpdateModifierRoutes(),
		&handlers.UpdateModifierHandler{Service: svc},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteModifierRequest, handlers.DeleteModifierResponse](
		ctx, router, handlers.DeleteModifierRoutes(),
		&handlers.DeleteModifierHandler{Service: svc},
	)
}
