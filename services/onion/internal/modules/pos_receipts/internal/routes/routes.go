package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_receipts/internal/application"
	"github.com/voxire/onion/internal/modules/pos_receipts/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGETWithValidation[handlers.ListReceiptsRequest, handlers.ListReceiptsResponse](
		ctx, router, handlers.ListReceiptsRoutes(),
		&handlers.ListReceiptsHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.EmailReceiptRequest, handlers.EmailReceiptResponse](
		ctx, router, handlers.EmailReceiptRoutes(),
		&handlers.EmailReceiptHandler{Service: svc},
	)
}
