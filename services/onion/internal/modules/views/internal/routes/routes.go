package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/views/internal/application"
	"github.com/voxire/onion/internal/modules/views/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGETWithValidation[handlers.ListViewsRequest, handlers.ListViewsResponse](
		ctx, router, handlers.ViewListRoutes(),
		&handlers.ListViewsHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CreateViewRequest, handlers.CreateViewResponse](
		ctx, router, handlers.ViewCreateRoutes(),
		&handlers.CreateViewHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[handlers.UpdateViewRequest, handlers.UpdateViewResponse](
		ctx, router, handlers.ViewUpdateRoutes(),
		&handlers.UpdateViewHandler{Service: svc},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteViewRequest, handlers.DeleteViewResponse](
		ctx, router, handlers.ViewDeleteRoutes(),
		&handlers.DeleteViewHandler{Service: svc},
	)
}
