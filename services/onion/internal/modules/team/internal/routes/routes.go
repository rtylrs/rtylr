package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/team/internal/application"
	"github.com/voxire/onion/internal/modules/team/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.ListUsersRequest, handlers.ListUsersResponse](
		ctx, router, handlers.ListUsersRoutes(),
		&handlers.ListUsersHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.InviteUserRequest, handlers.InviteUserResponse](
		ctx, router, handlers.InviteUserRoutes(),
		&handlers.InviteUserHandler{Service: service},
	)

	serverHttp.RegisterPUTWithValidation[handlers.UpdatePermissionsRequest, handlers.UpdatePermissionsResponse](
		ctx, router, handlers.UpdatePermissionsRoutes(),
		&handlers.UpdatePermissionsHandler{Service: service},
	)

	serverHttp.RegisterDELETEWithValidation[handlers.DeactivateUserRequest, handlers.DeactivateUserResponse](
		ctx, router, handlers.ListUsersRoutes(),
		&handlers.DeactivateUserHandler{Service: service},
	)

	serverHttp.RegisterGET[handlers.MeRequest, handlers.MeResponse](
		ctx, router, handlers.MeRoutes(),
		&handlers.MeHandler{Service: service},
	)

	// Invite verification and acceptance (anonymous - no auth required)
	serverHttp.RegisterGET[handlers.VerifyInviteRequest, handlers.VerifyInviteResponse](
		ctx, router, handlers.VerifyInviteRoutes(),
		&handlers.VerifyInviteHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.AcceptInviteRequest, handlers.AcceptInviteResponse](
		ctx, router, handlers.AcceptInviteRoutes(),
		&handlers.AcceptInviteHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.RequestAccessRequest, handlers.RequestAccessResponse](
		ctx, router, handlers.RequestAccessRoutes(),
		&handlers.RequestAccessHandler{Service: service},
	)
}
