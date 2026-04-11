package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/support_knowledge/internal/application"
	"github.com/voxire/onion/internal/modules/support_knowledge/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListArticlesRequest, handlers.ListArticlesResponse](
		ctx, router, handlers.ListArticlesRoutes(),
		&handlers.ListArticlesHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.CreateArticleRequest, handlers.CreateArticleResponse](
		ctx, router, handlers.CreateArticleRoutes(),
		&handlers.CreateArticleHandler{Service: service},
	)

	http.RegisterPUTWithValidation[handlers.UpdateArticleRequest, handlers.UpdateArticleResponse](
		ctx, router, handlers.UpdateArticleRoutes(),
		&handlers.UpdateArticleHandler{Service: service},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteArticleRequest, handlers.DeleteArticleResponse](
		ctx, router, handlers.DeleteArticleRoutes(),
		&handlers.DeleteArticleHandler{Service: service},
	)
}
