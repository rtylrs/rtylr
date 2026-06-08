package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/modules/insights/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func registerSettingsRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	serverHttp.RegisterGET[handlers.GetSettingsRequest, handlers.GetSettingsResponse](
		ctx, router, handlers.GetSettingsRoutes(),
		&handlers.GetSettingsHandler{Service: svc},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateSettingsRequest, handlers.UpdateSettingsResponse](
		ctx, router, handlers.UpdateSettingsRoutes(),
		&handlers.UpdateSettingsHandler{Service: svc},
	)
	serverHttp.RegisterGET[handlers.ListScheduledReportsRequest, handlers.ListScheduledReportsResponse](
		ctx, router, handlers.ListScheduledReportsRoutes(),
		&handlers.ListScheduledReportsHandler{Service: svc},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateScheduledReportRequest, handlers.CreateScheduledReportResponse](
		ctx, router, handlers.CreateScheduledReportRoutes(),
		&handlers.CreateScheduledReportHandler{Service: svc},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateScheduledReportRequest, handlers.UpdateScheduledReportResponse](
		ctx, router, handlers.UpdateScheduledReportRoutes(),
		&handlers.UpdateScheduledReportHandler{Service: svc},
	)
	serverHttp.RegisterDELETEWithValidation[handlers.DeleteScheduledReportRequest, handlers.DeleteScheduledReportResponse](
		ctx, router, handlers.DeleteScheduledReportRoutes(),
		&handlers.DeleteScheduledReportHandler{Service: svc},
	)
}
