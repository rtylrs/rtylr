package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/payroll/internal/application"
	"github.com/voxire/onion/internal/modules/payroll/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListPayrollRequest, handlers.ListPayrollResponse](
		ctx,
		router,
		handlers.ListPayrollRoutes(),
		&handlers.ListPayrollHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.RunPayrollRequest, handlers.RunPayrollResponse](
		ctx,
		router,
		handlers.RunPayrollRoutes(),
		&handlers.RunPayrollHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.MarkPayrollPaidRequest, handlers.MarkPayrollPaidResponse](
		ctx,
		router,
		handlers.MarkPayrollPaidRoutes(),
		&handlers.MarkPayrollPaidHandler{Service: service},
	)
}
