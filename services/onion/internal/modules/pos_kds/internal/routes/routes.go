package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/pos_kds/internal/application"
	"github.com/voxire/onion/internal/modules/pos_kds/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.ListKDSOrdersRequest, handlers.ListKDSOrdersResponse](
		ctx, router, handlers.ListKDSOrdersRoutes(),
		&handlers.ListKDSOrdersHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[handlers.UpdateKDSItemRequest, handlers.UpdateKDSItemResponse](
		ctx, router, handlers.UpdateKDSItemRoutes(),
		&handlers.UpdateKDSItemHandler{Service: svc},
	)
}
