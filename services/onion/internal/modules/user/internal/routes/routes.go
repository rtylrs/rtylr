package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/user/internal/application"
	"github.com/voxire/onion/internal/modules/user/internal/routes/handlers"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	accountv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/shared/account/v1"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, userService *application.Service, provisionerResolver authPkg.PlatformProvisionerResolver) {
	http.RegisterPOSTWithValidation[accountv1.RegisterRequest, accountv1.RegisterResponse](
		ctx,
		router,
		handlers.RegisterRoutes(),
		&handlers.RegisterHandler{UserService: userService, ProvisionerResolver: provisionerResolver},
	)
	http.RegisterPOSTWithValidation[accountv1.LoginRequest, accountv1.LoginResponse](
		ctx,
		router,
		handlers.LoginRoutes(),
		&handlers.LoginHandler{UserService: userService, ProvisionerResolver: provisionerResolver},
	)

	http.RegisterGET[accountv1.GetProfileRequest, accountv1.GetProfileResponse](
		ctx,
		router,
		handlers.AccountRoutes(),
		&handlers.AccountHandler{UserService: userService},
	)

	http.RegisterPOSTWithValidation[accountv1.UpdateProfileRequest, accountv1.UpdateProfileResponse](
		ctx,
		router,
		handlers.UpdateProfileRoutes(),
		&handlers.UpdateProfileHandler{UserService: userService},
	)

	http.RegisterPOSTWithValidation[accountv1.ChangePasswordRequest, accountv1.ChangePasswordResponse](
		ctx,
		router,
		handlers.ChangePasswordRoutes(),
		&handlers.ChangePasswordHandler{UserService: userService},
	)

	http.RegisterPOSTWithValidation[accountv1.DeleteAccountRequest, accountv1.DeleteAccountResponse](
		ctx,
		router,
		handlers.DeleteAccountRoutes(),
		&handlers.DeleteAccountHandler{UserService: userService},
	)
}
