package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/pos_menu/internal/application"
	"github.com/voxire/onion/internal/modules/pos_menu/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.ListMenuItemsRequest, handlers.ListMenuItemsResponse](
		ctx, router, handlers.ListMenuItemsRoutes(),
		&handlers.ListMenuItemsHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CreateMenuItemRequest, handlers.CreateMenuItemResponse](
		ctx, router, handlers.CreateMenuItemRoutes(),
		&handlers.CreateMenuItemHandler{Service: svc},
	)
}
