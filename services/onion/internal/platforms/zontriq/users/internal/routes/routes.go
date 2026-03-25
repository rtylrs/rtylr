package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	usersv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/users/v1"
	http "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platforms/zontriq/users/internal/application"
	"github.com/voxire/onion/internal/platforms/zontriq/users/internal/routes/handlers"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, usersService *application.Service) {
	http.RegisterGET[usersv1.GetZontriqUserRequest, usersv1.GetZontriqUserResponse](
		ctx,
		router,
		handlers.GetZontriqUserRoutes(),
		&handlers.GetZontriqUserHandler{UsersService: usersService},
	)

	http.RegisterPOST[usersv1.CreateZontriqUserRequest, usersv1.CreateZontriqUserResponse](
		ctx,
		router,
		handlers.CreateZontriqUserRoutes(),
		&handlers.CreateZontriqUserHandler{UsersService: usersService},
	)

	http.RegisterPOST[usersv1.UpdatePreferencesRequest, usersv1.UpdatePreferencesResponse](
		ctx,
		router,
		handlers.UpdatePreferencesRoutes(),
		&handlers.UpdatePreferencesHandler{UsersService: usersService},
	)

	http.RegisterGET[usersv1.ListPlansRequest, usersv1.ListPlansResponse](
		ctx,
		router,
		handlers.ListPlansRoutes(),
		&handlers.ListPlansHandler{UsersService: usersService},
	)

	http.RegisterGET[usersv1.GetUserPlanRequest, usersv1.GetUserPlanResponse](
		ctx,
		router,
		handlers.GetUserPlanRoutes(),
		&handlers.GetUserPlanHandler{UsersService: usersService},
	)

	http.RegisterPOST[usersv1.CompleteOnboardingRequest, usersv1.CompleteOnboardingResponse](
		ctx,
		router,
		handlers.CompleteOnboardingRoutes(),
		&handlers.CompleteOnboardingHandler{UsersService: usersService},
	)

	http.RegisterPOST[handlers.WaitlistRequest, handlers.WaitlistResponse](
		ctx,
		router,
		handlers.WaitlistRoutes(),
		&handlers.WaitlistHandler{UsersService: usersService},
	)
}
