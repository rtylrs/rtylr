package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/hr_attendance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_attendance/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterPOSTWithValidation[handlers.ClockInRequest, handlers.ClockInResponse](
		ctx, router, handlers.ClockInRoutes(), &handlers.ClockInHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.ClockOutRequest, handlers.ClockOutResponse](
		ctx, router, handlers.ClockOutRoutes(), &handlers.ClockOutHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.GetAttendanceRequest, handlers.GetAttendanceResponse](
		ctx, router, handlers.GetAttendanceRoutes(), &handlers.GetAttendanceHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.GetLocationAttendanceRequest, handlers.GetLocationAttendanceResponse](
		ctx, router, handlers.GetLocationAttendanceRoutes(), &handlers.GetLocationAttendanceHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.GenerateQRRequest, handlers.GenerateQRResponse](
		ctx, router, handlers.GenerateQRRoutes(), &handlers.GenerateQRHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.SetClockInPINRequest, handlers.SetClockInPINResponse](
		ctx, router, handlers.SetClockInPINRoutes(), &handlers.SetClockInPINHandler{Service: service},
	)
}
