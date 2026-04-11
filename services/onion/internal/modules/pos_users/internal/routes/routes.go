package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/pos_users/internal/application"
	"github.com/voxire/onion/internal/modules/pos_users/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.ListPOSUsersRequest, handlers.ListPOSUsersResponse](
		ctx, router, handlers.POSUserListRoutes(), &handlers.ListPOSUsersHandler{Service: svc},
	)
	http.RegisterPOSTWithValidation[handlers.CreatePOSUserRequest, handlers.CreatePOSUserResponse](
		ctx, router, handlers.POSUserCreateRoutes(), &handlers.CreatePOSUserHandler{Service: svc},
	)
	http.RegisterPOSTWithValidation[handlers.UpdatePOSUserRequest, handlers.UpdatePOSUserResponse](
		ctx, router, handlers.POSUserUpdateRoutes(), &handlers.UpdatePOSUserHandler{Service: svc},
	)
	http.RegisterPOSTWithValidation[handlers.DeactivatePOSUserRequest, handlers.DeactivatePOSUserResponse](
		ctx, router, handlers.POSUserDeactivateRoutes(), &handlers.DeactivatePOSUserHandler{Service: svc},
	)
	http.RegisterPOSTWithValidation[handlers.VerifyPINRequest, handlers.VerifyPINResponse](
		ctx, router, handlers.POSUserVerifyPINRoutes(), &handlers.VerifyPINHandler{Service: svc},
	)
}
