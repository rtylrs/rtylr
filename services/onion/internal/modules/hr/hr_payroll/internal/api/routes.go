package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/application"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router serverHttp.LegacyRouter, service *application.Service) {
	// Payroll Config
	serverHttp.RegisterGET[GetPayrollConfigRequest, GetPayrollConfigResponse](
		ctx, router, GetPayrollConfigRoutes(), &GetPayrollConfigHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[UpsertPayrollConfigRequest, UpsertPayrollConfigResponse](
		ctx, router, UpsertPayrollConfigRoutes(), &UpsertPayrollConfigHandler{Service: service},
	)

	// Payroll Runs
	serverHttp.RegisterGET[ListPayrollRunsRequest, ListPayrollRunsResponse](
		ctx, router, ListPayrollRunsRoutes(), &ListPayrollRunsHandler{Service: service},
	)
	serverHttp.RegisterGET[GetPayrollRunRequest, GetPayrollRunResponse](
		ctx, router, GetPayrollRunRoutes(), &GetPayrollRunHandler{Service: service},
	)
	serverHttp.RegisterGET[ListMyPayrollRunsRequest, ListMyPayrollRunsResponse](
		ctx, router, ListMyPayrollRunsRoutes(), &ListMyPayrollRunsHandler{Service: service},
	)
	serverHttp.RegisterGET[GetMyPayrollRunRequest, GetMyPayrollRunResponse](
		ctx, router, GetMyPayrollRunRoutes(), &GetMyPayrollRunHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[TriggerPayrollRunRequest, TriggerPayrollRunResponse](
		ctx, router, TriggerPayrollRunRoutes(), &TriggerPayrollRunHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[ApprovePayrollRunRequest, ApprovePayrollRunResponse](
		ctx, router, ApprovePayrollRunRoutes(), &ApprovePayrollRunHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[ExportPayrollRequest, ExportPayrollResponse](
		ctx, router, ExportPayrollRoutes(), &ExportPayrollHandler{Service: service},
	)
}
