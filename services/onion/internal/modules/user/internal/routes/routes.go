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
	// Google auth shares the tighter auth-endpoint budget -- same
	// 5/min/IP as /login. Both the login-by-google and signup-by-google
	// branches land in this handler.
	http.RegisterPOSTWithValidationAndMiddleware[handlers.GoogleAuthRequest, handlers.GoogleAuthResponse](
		ctx,
		router,
		handlers.GoogleAuthRoutes(),
		&handlers.GoogleAuthHandler{UserService: userService, OnUserCreated: userCreatedCb, OnUserLogin: userLoginCb},
		http.PerRouteRateLimit(5, time.Minute),
	)
	// GET /auth/google/authorize -- returns the WorkOS Google OAuth consent URL.
	// No rate-limit beyond global; the response is just a redirect URL.
	http.RegisterGET[handlers.GoogleAuthorizeRequest, handlers.GoogleAuthorizeResponse](
		ctx,
		router,
		handlers.GoogleAuthorizeRoutes(),
		&handlers.GoogleAuthorizeHandler{UserService: userService},
	)

	// GET /get-email-exists?email= -- anonymous email-existence check that
	// drives inline register/login validation. Returns only {exists: bool};
	// rate-limited per IP since it is a user-enumeration surface.
	http.RegisterGETWithValidationAndMiddleware[handlers.EmailExistsRequest, handlers.EmailExistsResponse](
		ctx,
		router,
		handlers.EmailExistsRoutes(),
		&handlers.EmailExistsHandler{UserService: userService},
		http.PerRouteRateLimit(20, time.Minute),
	)

	// GET /get-business-exists?name= -- inline business-name availability check
	// for the onboarding "create your business" screen. Returns {exists: bool}.
	http.RegisterGETWithValidation[handlers.BusinessExistsRequest, handlers.BusinessExistsResponse](
		ctx,
		router,
		handlers.BusinessExistsRoutes(),
		&handlers.BusinessExistsHandler{UserService: userService},
	)

	// POST /onboarding/business -- authenticated; creates the owner's first
	// business (globally-unique name) and returns a fresh token with its business_id.
	http.RegisterPOSTWithValidation[handlers.OnboardBusinessRequest, handlers.OnboardBusinessResponse](
		ctx,
		router,
		handlers.OnboardBusinessRoutes(),
		&handlers.OnboardBusinessHandler{UserService: userService},
	)

	http.RegisterGET[handlers.GetProfileRequest, handlers.GetProfileResponse](
		ctx,
		router,
		handlers.AccountRoutes(),
		&handlers.AccountHandler{UserService: userService},
	)

	http.RegisterPATCHWithValidation[handlers.UpdateProfileRequest, handlers.UpdateProfileResponse](
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

	http.RegisterPOSTWithValidationAndMiddleware[handlers.PasswordResetRequestBody, handlers.PasswordResetRequestResponse](
		ctx,
		router,
		handlers.PasswordResetRequestRoutes(),
		&handlers.PasswordResetRequestHandler{UserService: userService},
		http.PerRouteRateLimit(3, time.Minute),
	)
	http.RegisterPOSTWithValidation[handlers.PasswordResetConfirmBody, handlers.PasswordResetConfirmResponse](
		ctx,
		router,
		handlers.PasswordResetConfirmRoutes(),
		&handlers.PasswordResetConfirmHandler{UserService: userService},
	)

	http.RegisterPOSTWithValidation[handlers.SwitchBusinessRequest, handlers.SwitchBusinessResponse](
		ctx,
		router,
		handlers.SwitchBusinessRoutes(),
		&handlers.SwitchBusinessHandler{UserService: userService},
	)
	http.RegisterGET[handlers.VerifyEmailRequest, handlers.VerifyEmailResponse](
		ctx,
		router,
		handlers.VerifyEmailTokenRoutes(),
		&handlers.VerifyEmailHandler{},
	)
	http.RegisterPOSTWithValidationAndMiddleware[handlers.VerifyEmailRequest, handlers.VerifyEmailResponse](
		ctx,
		router,
		handlers.VerifyEmailBodyRoutes(),
		&handlers.VerifyEmailHandler{},
		http.PerRouteRateLimit(10, time.Minute),
	)
	http.RegisterPOSTWithValidationAndMiddleware[handlers.ResendVerificationRequest, handlers.ResendVerificationResponse](
		ctx,
		router,
		handlers.ResendVerificationRoutes(),
		&handlers.ResendVerificationHandler{},
		http.PerRouteRateLimit(3, time.Minute),
	)
}
