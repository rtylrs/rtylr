package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/tasks/internal/application"
	"github.com/voxire/onion/internal/modules/tasks/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.ListTasksRequest, handlers.ListTasksResponse](
		ctx, router, handlers.TaskListRoutes(),
		&handlers.ListTasksHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CreateTaskRequest, handlers.CreateTaskResponse](
		ctx, router, handlers.TaskCreateRoutes(),
		&handlers.CreateTaskHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[handlers.UpdateTaskRequest, handlers.UpdateTaskResponse](
		ctx, router, handlers.TaskUpdateRoutes(),
		&handlers.UpdateTaskHandler{Service: svc},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteTaskRequest, handlers.DeleteTaskResponse](
		ctx, router, handlers.TaskDeleteRoutes(),
		&handlers.DeleteTaskHandler{Service: svc},
	)

	http.RegisterGET[handlers.ListLabelsRequest, handlers.ListLabelsResponse](
		ctx, router, handlers.LabelListRoutes(),
		&handlers.ListLabelsHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CreateLabelRequest, handlers.CreateLabelResponse](
		ctx, router, handlers.LabelCreateRoutes(),
		&handlers.CreateLabelHandler{Service: svc},
	)
}
