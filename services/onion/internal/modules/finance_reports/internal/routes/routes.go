package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/finance_reports/internal/application"
	"github.com/voxire/onion/internal/modules/finance_reports/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.BalanceSheetRequest, handlers.BalanceSheetResponse](
		ctx, router, handlers.BalanceSheetRoutes(), &handlers.BalanceSheetHandler{Service: service},
	)
	http.RegisterGETWithValidation[handlers.CashFlowRequest, handlers.CashFlowResponse](
		ctx, router, handlers.CashFlowRoutes(), &handlers.CashFlowHandler{Service: service},
	)
}
