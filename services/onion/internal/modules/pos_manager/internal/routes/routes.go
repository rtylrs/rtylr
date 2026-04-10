package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/pos_manager/internal/application"
	"github.com/voxire/onion/internal/modules/pos_manager/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterPOSTWithValidation[handlers.VerifyPINRequest, handlers.VerifyPINResponse](
		ctx, router, handlers.VerifyPINRoutes(),
		&handlers.VerifyPINHandler{Service: svc},
	)
}
