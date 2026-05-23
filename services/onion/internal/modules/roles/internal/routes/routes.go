package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/roles/internal/application"
	"github.com/voxire/onion/internal/modules/roles/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.ListRolesRequest, handlers.ListRolesResponse](
		ctx, router, handlers.RoleRoutes(),
		&handlers.ListRolesHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.CreateRoleRequest, handlers.CreateRoleResponse](
		ctx, router, handlers.RoleRoutes(),
		&handlers.CreateRoleHandler{Service: service},
	)

	serverHttp.RegisterPUTWithValidation[handlers.UpdateRoleRequest, handlers.UpdateRoleResponse](
		ctx, router, handlers.RoleRoutes(),
		&handlers.UpdateRoleHandler{Service: service},
	)

	serverHttp.RegisterDELETEWithValidation[handlers.DeleteRoleRequest, handlers.DeleteRoleResponse](
		ctx, router, handlers.RoleRoutes(),
		&handlers.DeleteRoleHandler{Service: service},
	)

	serverHttp.RegisterGET[handlers.ListPermissionsRequest, handlers.ListPermissionsResponse](
		ctx, router, handlers.PermissionRoutes(),
		&handlers.ListPermissionsHandler{Service: service},
	)
}
