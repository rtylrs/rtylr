package routes

import (
	"context"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/user/internal/application"
	"github.com/voxire/onion/internal/modules/user/internal/routes/handlers"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, userService *application.Service, provisionerResolver authPkg.PlatformProvisionerResolver, callbacks ...func(ctx context.Context, userID int64)) {
	var userCreatedCb func(ctx context.Context, userID int64)
	var userLoginCb func(ctx context.Context, userID int64)
	if len(callbacks) > 0 {
		userCreatedCb = callbacks[0]
	}
	if len(callbacks) > 1 {
		userLoginCb = callbacks[1]
	}

	// Auth endpoints get a tighter per-route limit on top of the
	// global 100/min. 3/min on register and 5/min on login are wide
	// enough for legitimate retries (wrong-password humans, reloads
	// during signup) and narrow enough to make credential-stuffing +
	// signup-spam impractical from a single IP.
	http.RegisterPOSTWithValidationAndMiddleware[handlers.RegisterRequest, handlers.RegisterResponse](
		ctx,
		router,
		handlers.RegisterRoutes(),
		&handlers.RegisterHandler{UserService: userService, ProvisionerResolver: provisionerResolver, OnUserCreated: userCreatedCb},
		http.PerRouteRateLimit(3, time.Minute),
	)
	http.RegisterPOSTWithValidationAndMiddleware[handlers.LoginRequest, handlers.LoginResponse](
		ctx,
		router,
		handlers.LoginRoutes(),
		&handlers.LoginHandler{UserService: userService, ProvisionerResolver: provisionerResolver, OnUserLogin: userLoginCb},
		http.PerRouteRateLimit(5, time.Minute),
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
