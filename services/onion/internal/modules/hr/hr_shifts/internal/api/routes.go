package api

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/application"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[ListShiftsRequest, ListShiftsResponse](
		ctx, router, ListShiftsRoutes(), &ListShiftsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[CreateShiftRequest, CreateShiftResponse](
		ctx, router, CreateShiftRoutes(), &CreateShiftHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[UpdateShiftRequest, UpdateShiftResponse](
		ctx, router, UpdateShiftRoutes(), &UpdateShiftHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[DeleteShiftRequest, DeleteShiftResponse](
		ctx, router, DeleteShiftRoutes(), &DeleteShiftHandler{Service: service},
	)
	serverHttp.RegisterGET[GetRosterRequest, GetRosterResponse](
		ctx, router, GetRosterRoutes(), &GetRosterHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[CreateShiftAssignmentRequest, CreateShiftAssignmentResponse](
		ctx, router, CreateShiftAssignmentRoutes(), &CreateShiftAssignmentHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[PublishRosterRequest, PublishRosterResponse](
		ctx, router, PublishRosterRoutes(), &PublishRosterHandler{Service: service},
	)
	serverHttp.RegisterGET[ListAvailabilityRequest, ListAvailabilityResponse](
		ctx, router, ListAvailabilityRoutes(), &ListAvailabilityHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[SetAvailabilityRequest, SetAvailabilityResponse](
		ctx, router, SetAvailabilityRoutes(), &SetAvailabilityHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[DeleteAvailabilityRequest, DeleteAvailabilityResponse](
		ctx, router, DeleteAvailabilityRoutes(), &DeleteAvailabilityHandler{Service: service},
	)
}
