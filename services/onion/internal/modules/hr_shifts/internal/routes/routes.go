package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_shifts/internal/application"
	"github.com/voxire/onion/internal/modules/hr_shifts/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.ListShiftsRequest, handlers.ListShiftsResponse](
		ctx, router, handlers.ListShiftsRoutes(), &handlers.ListShiftsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateShiftRequest, handlers.CreateShiftResponse](
		ctx, router, handlers.CreateShiftRoutes(), &handlers.CreateShiftHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateShiftRequest, handlers.UpdateShiftResponse](
		ctx, router, handlers.UpdateShiftRoutes(), &handlers.UpdateShiftHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[handlers.DeleteShiftRequest, handlers.DeleteShiftResponse](
		ctx, router, handlers.DeleteShiftRoutes(), &handlers.DeleteShiftHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.GetRosterRequest, handlers.GetRosterResponse](
		ctx, router, handlers.GetRosterRoutes(), &handlers.GetRosterHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateShiftAssignmentRequest, handlers.CreateShiftAssignmentResponse](
		ctx, router, handlers.CreateShiftAssignmentRoutes(), &handlers.CreateShiftAssignmentHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.PublishRosterRequest, handlers.PublishRosterResponse](
		ctx, router, handlers.PublishRosterRoutes(), &handlers.PublishRosterHandler{Service: service},
	)
}
