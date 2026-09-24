package api

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_documents/internal/application"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterPOSTWithValidation[UploadDocumentRequest, UploadDocumentResponse](
		ctx, router, UploadDocumentRoutes(), &UploadDocumentHandler{Service: service},
	)
	serverHttp.RegisterGET[ListDocumentsRequest, any](
		ctx, router, ListDocumentsRoutes(), &ListDocumentsHandler{Service: service},
	)
	serverHttp.RegisterGET[ListExpiringDocumentsRequest, any](
		ctx, router, ListExpiringDocumentsRoutes(), &ListExpiringDocumentsHandler{Service: service},
	)
	serverHttp.RegisterGET[GetDocumentRequest, GetDocumentResponse](
		ctx, router, GetDocumentRoutes(), &GetDocumentHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[DeleteDocumentRequest, DeleteDocumentResponse](
		ctx, router, DeleteDocumentRoutes(), &DeleteDocumentHandler{Service: service},
	)
}
