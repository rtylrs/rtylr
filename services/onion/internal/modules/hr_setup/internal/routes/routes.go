package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_setup/internal/application"
	"github.com/voxire/onion/internal/modules/hr_setup/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.GetSetupStatusRequest, handlers.GetSetupStatusResponse](
		ctx, router, handlers.GetSetupStatusRoutes(), &handlers.GetSetupStatusHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CompleteSetupStepRequest, handlers.CompleteSetupStepResponse](
		ctx, router, handlers.CompleteSetupStepRoutes(), &handlers.CompleteSetupStepHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.GetHRConfigRequest, handlers.GetHRConfigResponse](
		ctx, router, handlers.GetHRConfigRoutes(), &handlers.GetHRConfigHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpsertHRConfigRequest, handlers.UpsertHRConfigResponse](
		ctx, router, handlers.UpsertHRConfigRoutes(), &handlers.UpsertHRConfigHandler{Service: service},
	)
}
