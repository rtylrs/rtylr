package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	creditsv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/credits/v1"
	http "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platforms/zontriq/credits/internal/application"
	"github.com/voxire/onion/internal/platforms/zontriq/credits/internal/routes/handlers"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, creditsService *application.Service) {
	http.RegisterGET[creditsv1.GetCreditsRequest, creditsv1.GetCreditsResponse](
		ctx,
		router,
		handlers.GetCreditsRoutes(),
		&handlers.GetCreditsHandler{CreditsService: creditsService},
	)

	http.RegisterGET[creditsv1.ListCreditTransactionsRequest, creditsv1.ListCreditTransactionsResponse](
		ctx,
		router,
		handlers.ListTransactionsRoutes(),
		&handlers.ListTransactionsHandler{CreditsService: creditsService},
	)

	http.RegisterPOSTWithValidation[creditsv1.PurchaseCreditsRequest, creditsv1.PurchaseCreditsResponse](
		ctx,
		router,
		handlers.PurchaseCreditsRoutes(),
		&handlers.PurchaseCreditsHandler{CreditsService: creditsService},
	)

	http.RegisterPOSTWithValidation[creditsv1.SpendCreditsRequest, creditsv1.SpendCreditsResponse](
		ctx,
		router,
		handlers.SpendCreditsRoutes(),
		&handlers.SpendCreditsHandler{CreditsService: creditsService},
	)
}
