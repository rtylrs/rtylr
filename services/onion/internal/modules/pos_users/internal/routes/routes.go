package routes

import (
	"context"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_users/internal/application"
	"github.com/voxire/onion/internal/modules/pos_users/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

// verifyPINRateLimit: 5 attempts per minute per IP on /pos/users/verify-pin.
// This stacks on top of the global rate limiter and is intended to slow
// down automated PIN-guessing against a known display name. The numbers
// are intentionally tight because a POS PIN is only 4–10 digits.
const (
	verifyPINMaxPerMinute = 5
	verifyPINWindow       = time.Minute
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.ListPOSUsersRequest, handlers.ListPOSUsersResponse](
		ctx, router, handlers.POSUserListRoutes(), &handlers.ListPOSUsersHandler{Service: svc},
	)
	http.RegisterPOSTWithValidation[handlers.CreatePOSUserRequest, handlers.CreatePOSUserResponse](
		ctx, router, handlers.POSUserCreateRoutes(), &handlers.CreatePOSUserHandler{Service: svc},
	)
	http.RegisterPATCHWithValidation[handlers.UpdatePOSUserRequest, handlers.UpdatePOSUserResponse](
		ctx, router, handlers.POSUserUpdateRoutes(), &handlers.UpdatePOSUserHandler{Service: svc},
	)
	http.RegisterPATCHWithValidation[handlers.DeactivatePOSUserRequest, handlers.DeactivatePOSUserResponse](
		ctx, router, handlers.POSUserDeactivateRoutes(), &handlers.DeactivatePOSUserHandler{Service: svc},
	)
	// verify-pin is validation-gated AND rate-limited. Main already
	// has `RegisterPOSTWithValidationAndMiddleware` + 2-arg
	// `PerRouteRateLimit` from the /auth/login work earlier — reusing
	// both so there's a single rate-limiter pattern across the repo.
	http.RegisterPOSTWithValidationAndMiddleware[handlers.VerifyPINRequest, handlers.VerifyPINResponse](
		ctx, router, handlers.POSUserVerifyPINRoutes(), &handlers.VerifyPINHandler{Service: svc},
		http.PerRouteRateLimit(verifyPINMaxPerMinute, verifyPINWindow),
	)
	http.RegisterPOSTWithValidation[handlers.SetPINRequest, handlers.SetPINResponse](
		ctx, router, handlers.POSUserSetPINRoutes(), &handlers.SetPINHandler{Service: svc},
	)
}
