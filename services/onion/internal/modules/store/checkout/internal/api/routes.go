package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/store/checkout/internal/application"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router http.LegacyRouter, service *application.Service) {
	http.RegisterPOSTWithValidation[CreateRequest, CreateResponse](ctx, router, CreateRoutes(), &CreateHandler{Service: service})
	http.RegisterPOSTWithValidation[CompleteRequest, CompleteResponse](ctx, router, CompleteRoutes(), &CompleteHandler{Service: service})
}
