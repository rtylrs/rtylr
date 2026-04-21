package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/modules/insights/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func registerCustomerRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	serverHttp.RegisterGET[handlers.GetCustomerKPIsRequest, handlers.GetCustomerKPIsResponse](
		ctx, router, handlers.GetCustomerKPIsRoutes(),
		&handlers.GetCustomerKPIsHandler{Service: svc},
	)
	serverHttp.RegisterGET[handlers.GetTopCustomersRequest, handlers.GetTopCustomersResponse](
		ctx, router, handlers.GetTopCustomersRoutes(),
		&handlers.GetTopCustomersHandler{Service: svc},
	)
	serverHttp.RegisterGET[handlers.GetLTVRequest, handlers.GetLTVResponse](
		ctx, router, handlers.GetLTVRoutes(),
		&handlers.GetLTVHandler{Service: svc},
	)
}
