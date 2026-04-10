package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/pos_gift_cards/internal/application"
	"github.com/voxire/onion/internal/modules/pos_gift_cards/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.ListGiftCardsRequest, handlers.ListGiftCardsResponse](
		ctx, router, handlers.ListGiftCardsRoutes(),
		&handlers.ListGiftCardsHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CreateGiftCardRequest, handlers.CreateGiftCardResponse](
		ctx, router, handlers.CreateGiftCardRoutes(),
		&handlers.CreateGiftCardHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[handlers.RedeemGiftCardRequest, handlers.RedeemGiftCardResponse](
		ctx, router, handlers.RedeemGiftCardRoutes(),
		&handlers.RedeemGiftCardHandler{Service: svc},
	)
}
