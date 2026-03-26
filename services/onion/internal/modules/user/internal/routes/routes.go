package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/user/internal/application"
	"github.com/voxire/onion/internal/modules/user/internal/routes/handlers"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, userService *application.Service, provisionerResolver authPkg.PlatformProvisionerResolver) {
	http.RegisterPOSTWithValidation[handlers.RegisterRequest, handlers.RegisterResponse](
		ctx,
		router,
		handlers.RegisterRoutes(),
		&handlers.RegisterHandler{UserService: userService, ProvisionerResolver: provisionerResolver},
	)
	http.RegisterPOSTWithValidation[handlers.LoginRequest, handlers.LoginResponse](
		ctx,
		router,
		handlers.LoginRoutes(),
		&handlers.LoginHandler{UserService: userService, ProvisionerResolver: provisionerResolver},
	)

	http.RegisterGET[handlers.GetProfileRequest, handlers.GetProfileResponse](
		ctx,
		router,
		handlers.AccountRoutes(),
		&handlers.AccountHandler{UserService: userService},
	)

	http.RegisterPOSTWithValidation[handlers.UpdateProfileRequest, handlers.UpdateProfileResponse](
		ctx,
		router,
		handlers.UpdateProfileRoutes(),
		&handlers.UpdateProfileHandler{UserService: userService},
	)

	http.RegisterPOSTWithValidation[handlers.ChangePasswordRequest, handlers.ChangePasswordResponse](
		ctx,
		router,
		handlers.ChangePasswordRoutes(),
		&handlers.ChangePasswordHandler{UserService: userService},
	)

	http.RegisterPOSTWithValidation[handlers.DeleteAccountRequest, handlers.DeleteAccountResponse](
		ctx,
		router,
		handlers.DeleteAccountRoutes(),
		&handlers.DeleteAccountHandler{UserService: userService},
	)
}
