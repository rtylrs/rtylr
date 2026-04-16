package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/hr_overtime/internal/application"
	"github.com/voxire/onion/internal/modules/hr_overtime/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.ListOvertimeRequestsRequest, handlers.ListOvertimeRequestsResponse](
		ctx, router, handlers.ListOvertimeRequestsRoutes(), &handlers.ListOvertimeRequestsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateOvertimeRequestRequest, handlers.CreateOvertimeRequestResponse](
		ctx, router, handlers.CreateOvertimeRequestRoutes(), &handlers.CreateOvertimeRequestHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.ApproveOvertimeRequestReq, handlers.ApproveOvertimeRequestRes](
		ctx, router, handlers.ApproveOvertimeRoutes(), &handlers.ApproveOvertimeHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.GetOvertimeConfigRequest, handlers.GetOvertimeConfigResponse](
		ctx, router, handlers.GetOvertimeConfigRoutes(), &handlers.GetOvertimeConfigHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpsertOvertimeConfigRequest, handlers.UpsertOvertimeConfigResponse](
		ctx, router, handlers.UpsertOvertimeConfigRoutes(), &handlers.UpsertOvertimeConfigHandler{Service: service},
	)
}
