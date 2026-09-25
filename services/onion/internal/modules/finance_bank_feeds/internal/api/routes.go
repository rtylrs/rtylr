package api

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/finance_bank_feeds/internal/application"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[ListRequest, ListResponse](
		ctx, router, ListRoutes(), &ListHandler{Service: service},
	)
	http.RegisterPOSTWithValidation[SyncRequest, SyncResponse](
		ctx, router, SyncRoutes(), &SyncHandler{Service: service},
	)
}
