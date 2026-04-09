package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/projects/internal/application"
	"github.com/voxire/onion/internal/modules/projects/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.ListProjectsRequest, handlers.ListProjectsResponse](
		ctx, router, handlers.ProjectListRoutes(),
		&handlers.ListProjectsHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CreateProjectRequest, handlers.CreateProjectResponse](
		ctx, router, handlers.ProjectCreateRoutes(),
		&handlers.CreateProjectHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[handlers.UpdateProjectRequest, handlers.UpdateProjectResponse](
		ctx, router, handlers.ProjectUpdateRoutes(),
		&handlers.UpdateProjectHandler{Service: svc},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteProjectRequest, handlers.DeleteProjectResponse](
		ctx, router, handlers.ProjectDeleteRoutes(),
		&handlers.DeleteProjectHandler{Service: svc},
	)
}
