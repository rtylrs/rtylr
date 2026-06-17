package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_tables/internal/application"
	"github.com/voxire/onion/internal/modules/pos_tables/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.ListTablesRequest, handlers.ListTablesResponse](
		ctx, router, handlers.TableListRoutes(),
		&handlers.ListTablesHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CreateTableRequest, handlers.CreateTableResponse](
		ctx, router, handlers.TableCreateRoutes(),
		&handlers.CreateTableHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[handlers.UpdateTableRequest, handlers.UpdateTableResponse](
		ctx, router, handlers.TableUpdateRoutes(),
		&handlers.UpdateTableHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[handlers.UpdateTableStatusRequest, handlers.UpdateTableStatusResponse](
		ctx, router, handlers.TableUpdateStatusRoutes(),
		&handlers.UpdateTableStatusHandler{Service: svc},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteTableRequest, handlers.DeleteTableResponse](
		ctx, router, handlers.TableDeleteRoutes(),
		&handlers.DeleteTableHandler{Service: svc},
	)
}
