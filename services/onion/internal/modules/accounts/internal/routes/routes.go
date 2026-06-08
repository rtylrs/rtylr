package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/accounts/internal/application"
	"github.com/voxire/onion/internal/modules/accounts/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListRequest, handlers.ListResponse](
		ctx, router, handlers.ListRoutes(), &handlers.ListHandler{Service: service},
	)
	http.RegisterPOSTWithValidation[handlers.CreateRequest, handlers.CreateResponse](
		ctx, router, handlers.CreateRoutes(), &handlers.CreateHandler{Service: service},
	)
	http.RegisterPUTWithValidation[handlers.UpdateAccountRequest, handlers.UpdateAccountResponse](
		ctx, router, handlers.UpdateAccountRoutes(), &handlers.UpdateAccountHandler{Service: service},
	)
	http.RegisterDELETEWithValidation[handlers.DeleteAccountRequest, handlers.DeleteAccountResponse](
		ctx, router, handlers.DeleteAccountRoutes(), &handlers.DeleteAccountHandler{Service: service},
	)
	http.RegisterGET[handlers.ChartOfAccountsRequest, handlers.ChartOfAccountsResponse](
		ctx, router, handlers.ChartOfAccountsRoutes(),
		&handlers.ChartOfAccountsHandler{Service: service},
	)
}
