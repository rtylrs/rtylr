package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/application"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	// Leave Types
	serverHttp.RegisterGET[handlers.ListLeaveTypesRequest, handlers.ListLeaveTypesResponse](
		ctx, router, handlers.ListLeaveTypesRoutes(), &handlers.ListLeaveTypesHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateLeaveTypeRequest, handlers.CreateLeaveTypeResponse](
		ctx, router, handlers.CreateLeaveTypeRoutes(), &handlers.CreateLeaveTypeHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateLeaveTypeRequest, handlers.UpdateLeaveTypeResponse](
		ctx, router, handlers.UpdateLeaveTypeRoutes(), &handlers.UpdateLeaveTypeHandler{Service: service},
	)

	// Leave Requests
	serverHttp.RegisterGET[handlers.ListLeaveRequestsRequest, handlers.ListLeaveRequestsResponse](
		ctx, router, handlers.ListLeaveRequestsRoutes(), &handlers.ListLeaveRequestsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateLeaveRequestRequest, handlers.CreateLeaveRequestResponse](
		ctx, router, handlers.CreateLeaveRequestRoutes(), &handlers.CreateLeaveRequestHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.ApproveLeaveRequest, handlers.ApproveLeaveResponse](
		ctx, router, handlers.ApproveLeaveRoutes(), &handlers.ApproveLeaveHandler{Service: service},
	)

	// Leave Balances
	serverHttp.RegisterGET[handlers.GetLeaveBalancesRequest, handlers.GetLeaveBalancesResponse](
		ctx, router, handlers.GetLeaveBalancesRoutes(), &handlers.GetLeaveBalancesHandler{Service: service},
	)

	// Leave Calendar
	serverHttp.RegisterGET[handlers.GetLeaveCalendarRequest, handlers.GetLeaveCalendarResponse](
		ctx, router, handlers.GetLeaveCalendarRoutes(), &handlers.GetLeaveCalendarHandler{Service: service},
	)
}
