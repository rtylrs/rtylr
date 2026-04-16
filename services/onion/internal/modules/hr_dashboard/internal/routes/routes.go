package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/hr_dashboard/internal/application"
	"github.com/voxire/onion/internal/modules/hr_dashboard/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.GetDashboardWidgetsRequest, handlers.GetDashboardWidgetsResponse](
		ctx, router, handlers.GetDashboardWidgetsRoutes(), &handlers.GetDashboardWidgetsHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateDashboardWidgetsRequest, handlers.UpdateDashboardWidgetsResponse](
		ctx, router, handlers.UpdateDashboardWidgetsRoutes(), &handlers.UpdateDashboardWidgetsHandler{Service: service},
	)
}
