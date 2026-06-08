package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_payroll/internal/application"
	"github.com/voxire/onion/internal/modules/hr_payroll/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	// Payroll Config
	serverHttp.RegisterGET[handlers.GetPayrollConfigRequest, handlers.GetPayrollConfigResponse](
		ctx, router, handlers.GetPayrollConfigRoutes(), &handlers.GetPayrollConfigHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpsertPayrollConfigRequest, handlers.UpsertPayrollConfigResponse](
		ctx, router, handlers.UpsertPayrollConfigRoutes(), &handlers.UpsertPayrollConfigHandler{Service: service},
	)

	// Payroll Runs
	serverHttp.RegisterGET[handlers.ListPayrollRunsRequest, handlers.ListPayrollRunsResponse](
		ctx, router, handlers.ListPayrollRunsRoutes(), &handlers.ListPayrollRunsHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.GetPayrollRunRequest, handlers.GetPayrollRunResponse](
		ctx, router, handlers.GetPayrollRunRoutes(), &handlers.GetPayrollRunHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.TriggerPayrollRunRequest, handlers.TriggerPayrollRunResponse](
		ctx, router, handlers.TriggerPayrollRunRoutes(), &handlers.TriggerPayrollRunHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.ApprovePayrollRunRequest, handlers.ApprovePayrollRunResponse](
		ctx, router, handlers.ApprovePayrollRunRoutes(), &handlers.ApprovePayrollRunHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.ExportPayrollRequest, handlers.ExportPayrollResponse](
		ctx, router, handlers.ExportPayrollRoutes(), &handlers.ExportPayrollHandler{Service: service},
	)
}
