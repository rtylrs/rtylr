package api

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/store_checkout/internal/application"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterPOSTWithValidation[CreateRequest, CreateResponse](ctx, router, CreateRoutes(), &CreateHandler{Service: service})
	http.RegisterPOSTWithValidation[CompleteRequest, CompleteResponse](ctx, router, CompleteRoutes(), &CompleteHandler{Service: service})
}
