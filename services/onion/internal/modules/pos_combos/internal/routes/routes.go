package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_combos/internal/application"
	"github.com/voxire/onion/internal/modules/pos_combos/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.ListCombosRequest, handlers.ListCombosResponse](
		ctx, router, handlers.ListCombosRoutes(),
		&handlers.ListCombosHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CreateComboRequest, handlers.CreateComboResponse](
		ctx, router, handlers.CreateComboRoutes(),
		&handlers.CreateComboHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[handlers.UpdateComboRequest, handlers.UpdateComboResponse](
		ctx, router, handlers.UpdateComboRoutes(),
		&handlers.UpdateComboHandler{Service: svc},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteComboRequest, handlers.DeleteComboResponse](
		ctx, router, handlers.DeleteComboRoutes(),
		&handlers.DeleteComboHandler{Service: svc},
	)
}
