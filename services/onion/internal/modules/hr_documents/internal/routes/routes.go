package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_documents/internal/application"
	"github.com/voxire/onion/internal/modules/hr_documents/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterPOSTWithValidation[handlers.UploadDocumentRequest, handlers.UploadDocumentResponse](
		ctx, router, handlers.UploadDocumentRoutes(), &handlers.UploadDocumentHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.ListDocumentsRequest, handlers.ListDocumentsResponse](
		ctx, router, handlers.ListDocumentsRoutes(), &handlers.ListDocumentsHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.ListExpiringDocumentsRequest, handlers.ListDocumentsResponse](
		ctx, router, handlers.ListExpiringDocumentsRoutes(), &handlers.ListExpiringDocumentsHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.GetDocumentRequest, handlers.GetDocumentResponse](
		ctx, router, handlers.GetDocumentRoutes(), &handlers.GetDocumentHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[handlers.DeleteDocumentRequest, handlers.DeleteDocumentResponse](
		ctx, router, handlers.DeleteDocumentRoutes(), &handlers.DeleteDocumentHandler{Service: service},
	)
}
