package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/tax/internal/application"
	"github.com/voxire/onion/internal/modules/tax/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.GetTaxConfigRequest, handlers.GetTaxConfigResponse](
		ctx, router, handlers.TaxConfigGetRoutes(),
		&handlers.GetTaxConfigHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[handlers.UpdateTaxConfigRequest, handlers.UpdateTaxConfigResponse](
		ctx, router, handlers.TaxConfigUpdateRoutes(),
		&handlers.UpdateTaxConfigHandler{Service: svc},
	)

	http.RegisterGET[handlers.GetPLReportRequest, handlers.PLReportResponse](
		ctx, router, handlers.PLReportRoutes(),
		&handlers.GetPLReportHandler{Service: svc},
	)
}
